package main

import (
	"fmt"
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
	err := rollbar.WrapAndWait(http.ListenAndServe(":5000", handlers.CompressHandler(r)))

	if err == nil {
		log.Error(err)
	}
}

func NameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := fmt.Sprintf("%s/%s", vars["namespace"], vars["name"])
	var version, name, api, stacks string

	if version = vars["version"]; version != "" {
		version = "0.1"
	}

	if name = vars["name"]; name != "" {
		name = id
	}

	if api = vars["api"]; api != "" {
		api = "0.4"
	}

	if stacks = vars["stacks"]; stacks != "" {
		stack := vars["stack"]
		if stack != "" {
			stacks = stack
		} else {
			stacks = "heroku-18,heroku-20"
		}
	}

	shimDir, _ := os.Getwd()

	shimmedBuildpack := fmt.Sprintf("%s.tgz", uuid.New())
	dir, _ := os.Getwd()
	dir, _ = ioutil.TempDir(dir, uuid.New().String())
	defer os.RemoveAll(dir)
	_ = os.Chdir(dir)

	log.Infof("at=shim file=%s\n\n", shimmedBuildpack)

	_ = os.Mkdir("bin", 0777)
	input, _ := ioutil.ReadFile(fmt.Sprintf("%s/bin/build", shimDir))
	_ = ioutil.WriteFile("bin/build", input, 0644)

	input, _ = ioutil.ReadFile(fmt.Sprintf("%s/bin/detect", shimDir))
	_ = ioutil.WriteFile("bin/detect", input, 0644)

	input, _ = ioutil.ReadFile(fmt.Sprintf("%s/bin/release", shimDir))
	_ = ioutil.WriteFile("bin/release", input, 0644)

	input, _ = ioutil.ReadFile(fmt.Sprintf("%s/bin/exports", shimDir))
	_ = ioutil.WriteFile("bin/exports", input, 0644)

	log.Infof("at=descriptor file=%s api=%s id=%s version=%s name=%s stacks=%s\n\n",
		shimmedBuildpack, api, id, version, name, stacks)

	file, _ := os.Create("buildpack.toml")

	bp := fmt.Sprintf("api = %s\n\n[buildpack] id = %s\nversion = %s\nname = %s\n", api, id, version, name)

	for _, s := range strings.Split(stacks, ",") {
		s = fmt.Sprintf("\n[[stacks]]\nid = %s\n", s)
		bp = bp + s
	}

	_, _ = file.WriteString(bp)
	target_dir := "target"
	_ = os.Mkdir(target_dir, 0777)

	url := fmt.Sprintf("https://buildpack-registry.s3.amazonaws.com/buildpacks/%s.tgz", id)
	log.Infof("at=download file=%s url=%s\n\n", shimmedBuildpack, url)
	cmd := fmt.Sprintf(`curl --retry 3 --silent --location "%s" | tar xzm -C %s`, url, target_dir)

	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Error(err)
	}

	_ = os.Chdir(shimDir)
	cmd = fmt.Sprintf("tar cz --file=%s --directory=%s .", shimmedBuildpack, dir)

	_, err = exec.Command("bash", "-c", cmd).Output()
	defer fmt.Printf("at=cleanup file=%s\n\n", shimmedBuildpack)
	defer os.Remove(shimmedBuildpack)

	if err != nil {
		log.Error(err)
	}

	fstat, _ := file.Stat()
	log.Infof("at=send file=%s size=%d", shimmedBuildpack, fstat.Size())
	http.ServeFile(w, r, shimmedBuildpack)
	log.Infof("at=success file=%s", shimmedBuildpack)
}
