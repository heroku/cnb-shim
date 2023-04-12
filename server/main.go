package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/heroku/cnb-shim/config"
	"github.com/heroku/rollrus"
	"github.com/rollbar/rollbar-go"
	log "github.com/sirupsen/logrus"
)

func main() {
	var conf config.Config
	config.LoadConfig(&conf)
	rollrus.SetupLogging(conf.RollbarAccessToken, conf.RollbarEnvironment)

	r := mux.NewRouter()
	r.HandleFunc("/v1/{namespace}/{name}", NameHandler)
	r.HandleFunc("/health", HealthHandler)
	r.Use(PanicRecoveryMiddleware)

	port := fmt.Sprintf(":%s", conf.Port)
	http.ListenAndServe(port, handlers.CompressHandler(r))

	log.Info("server started")
	rollbar.Wait()
}

// PanicRecoveryMiddleware recovers from a panic that may have occurred during a
// request, reports the error to Rollbar, and sends back a 500.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recoverVal := recover(); recoverVal != nil {
				var err error
				var ok bool
				if err, ok = recoverVal.(error); ok {
					rollbar.LogPanic(err, false)
				}

				w.WriteHeader(http.StatusInternalServerError)
			}

		}()

		next.ServeHTTP(w, r)
	})
}

func NameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := fmt.Sprintf("%s/%s", vars["namespace"], vars["name"])
	var version, name, api, stacks string
	params := r.URL.Query()

	if version = params.Get("version"); version == "" {
		version = "0.1"
	}

	if name = params.Get("name"); name == "" {
		name = "0.1"
	}

	if api = params.Get("api"); api == "" {
		api = "0.8"
	}

	if stacks = params.Get("stacks"); stacks == "" {
		if stack := params.Get("stack"); stack == "" {
			stacks = "heroku-18,heroku-20,heroku-22"
		} else {
			stacks = stack
		}
	}

	appDir, _ := os.Getwd()

	shimmedBuildpack := fmt.Sprintf("%s/%s.tgz", appDir, uuid.New())
	dir, err := ioutil.TempDir("", uuid.New().String())
	handlePanic(err)
	defer os.RemoveAll(dir)

	log.Infof("at=shim file=%s", shimmedBuildpack)

	handlePanic(os.Mkdir(fmt.Sprintf("%s/bin/", dir), 0777))

	files := []string{"build", "detect", "release", "exports"}
	for _, f := range files {
		input, err := ioutil.ReadFile(fmt.Sprintf("%s/bin/%s", appDir, f))
		handlePanic(err)
		err = ioutil.WriteFile(fmt.Sprintf("%s/bin/%s", dir, f), input, 0700)
		handlePanic(err)
	}

	log.Infof("at=descriptor file=%s api=%s id=%s version=%s name=%s stacks=%s",
		shimmedBuildpack, api, id, version, name, stacks)

	file, err := os.Create(fmt.Sprintf("%s/buildpack.toml", dir))
	handlePanic(err)

	bp := fmt.Sprintf("api = \"%s\"\n\n[buildpack]\nid = \"%s\"\nversion = \"%s\"\nname = \"%s\"\n", api, id, version, name)

	for _, s := range strings.Split(stacks, ",") {
		s = fmt.Sprintf("\n[[stacks]]\nid = \"%s\"\n", s)
		bp = bp + s
	}

	_, err = file.WriteString(bp)
	handlePanic(err)

	target_dir := fmt.Sprintf("%s/target", dir)
	handlePanic(os.Mkdir(target_dir, 0777))

	url := fmt.Sprintf("https://buildpack-registry.s3.amazonaws.com/buildpacks/%s.tgz", id)
	log.Infof("at=download file=%s url=%s", shimmedBuildpack, url)
	bp, err = downloadBuildpack(url)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handlePanic(err)
	log.Infof("at=tar file=%s target_dir=%s", shimmedBuildpack, target_dir)
	tar := fmt.Sprintf(`tar xzf %s -C %s`, bp, target_dir)
	_, err = exec.Command("bash", "-c", tar).Output()
	handlePanic(err)
	handlePanic(os.Remove(bp))

	// The sort, mtime and owner options ensure the archive is deterministic across time and dynos
	// (since the user ID varies by dyno). See: https://reproducible-builds.org/docs/archives/
	// This helps reduce layer churn in builder images containing shimmed buildpacks.
	cmd := fmt.Sprintf("tar -cz --file=%s --directory=%s --sort=name --mtime='1980-01-01 00:00:01Z' --owner=0 --group=0 --numeric-owner .", shimmedBuildpack, dir)

	_, err = exec.Command("bash", "-c", cmd).Output()
	handlePanic(err)
	defer log.Infof("at=cleanup file=%s", shimmedBuildpack)
	defer os.Remove(shimmedBuildpack)

	fstat, err := file.Stat()
	handlePanic(err)
	log.Infof("at=send file=%s size=%d", shimmedBuildpack, fstat.Size())
	w.Header().Add("Content-Type", "application/x-gzip")
	http.ServeFile(w, r, shimmedBuildpack)
	log.Infof("at=success file=%s", shimmedBuildpack)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "health check ok")
	log.Info("health check ok")
}

func handlePanic(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func downloadBuildpack(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", "*.tgz")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(file, resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Failed to download buildpack %s (Status: %s)", url, resp.Status)
	}

	return file.Name(), err
}
