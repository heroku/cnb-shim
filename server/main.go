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

	port := fmt.Sprintf(":%s", conf.Port)
	rollbar.WrapAndWait(http.ListenAndServe(port, handlers.CompressHandler(r)))

	log.Info("server started")
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
		api = "0.4"
	}

	if stacks = params.Get("stacks"); stacks == "" {
		if stack := params.Get("stack"); stack == "" {
			stacks = "heroku-18,heroku-20"
		} else {
			stacks = stack
		}
	}

	shimDir, _ := os.Getwd()

	shimmedBuildpack := fmt.Sprintf("%s.tgz", uuid.New())
	dir, _ := os.Getwd()
	dir, _ = ioutil.TempDir(dir, uuid.New().String())
	defer os.RemoveAll(dir)
	handlePanic(os.Chdir(dir))

	log.Infof("at=shim file=%s", shimmedBuildpack)

	handlePanic(os.Mkdir("bin", 0777))

	files := []string{"build", "detect", "release", "exports"}
	for _, f := range files {
		input, err := ioutil.ReadFile(fmt.Sprintf("%s/bin/%s", shimDir, f))
		handlePanic(err)
		err = ioutil.WriteFile(fmt.Sprintf("bin/%s", f), input, 0700)
		handlePanic(err)
	}

	log.Infof("at=descriptor file=%s api=%s id=%s version=%s name=%s stacks=%s",
		shimmedBuildpack, api, id, version, name, stacks)

	file, err := os.Create("buildpack.toml")
	handlePanic(err)

	bp := fmt.Sprintf("api = \"%s\"\n\n[buildpack]\nid = \"%s\"\nversion = \"%s\"\nname = \"%s\"\n", api, id, version, name)

	for _, s := range strings.Split(stacks, ",") {
		s = fmt.Sprintf("\n[[stacks]]\nid = \"%s\"\n", s)
		bp = bp + s
	}

	_, err = file.WriteString(bp)
	handlePanic(err)
	target_dir := "target"
	handlePanic(os.Mkdir(target_dir, 0777))

	url := fmt.Sprintf("https://buildpack-registry.s3.amazonaws.com/buildpacks/%s.tgz", id)
	log.Infof("at=download file=%s url=%s", shimmedBuildpack, url)
	bp, err = downloadBuildpack(url)
	handlePanic(err)
	tar := fmt.Sprintf(`tar xzf %s -C %s`, bp, target_dir)
	_, err = exec.Command("bash", "-c", tar).Output()
	handlePanic(err)
	handlePanic(os.Remove(bp))
	handlePanic(os.Chdir(shimDir))

	cmd := fmt.Sprintf("tar cz --file=%s --directory=%s .", shimmedBuildpack, dir)

	_, err = exec.Command("bash", "-c", cmd).Output()
	handlePanic(err)
	defer fmt.Printf("at=cleanup file=%s", shimmedBuildpack)
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

	id := uuid.New().String() + ".tgz"
	out, err := os.Create(id)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, resp.Body)

	if err != nil {
		return "", err
	}

	return id, err
}
