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
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/{namespace}/{name}", NameHandler)
	err := http.ListenAndServe(":5000", handlers.CompressHandler(r))

	if err == nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := fmt.Sprintf("%s/%s", vars["namespace"], vars["name"])

	var version, name, api, stacks string
	var found bool

	if version, found = mux.Vars(r)["id"]; !found {
		version = "0.1"
	}

	if name, found = mux.Vars(r)["name"]; !found {
		name = id
	}

	if api, found = mux.Vars(r)["api"]; !found {
		api = "0.4"
	}

	if stacks, found = mux.Vars(r)["stacks"]; !found {
		stack, found := mux.Vars(r)["stacks"]
		if !found {
			stacks = "heroku-18,heroku-20"
		} else {
			stacks = stack
		}
	}

	shimDir, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	shimmedBuildpack := fmt.Sprintf("%s.tgz", uuid.New())
	dir, _ := os.Getwd()
	dir, _ = ioutil.TempDir(dir, uuid.New().String())
	defer os.RemoveAll(dir)
	_ = os.Chdir(dir)

	fmt.Printf("at=shim file=%s\n\n", shimmedBuildpack)

	input, _ := ioutil.ReadFile(fmt.Sprintf("%s/bin/build", shimDir))
	_ = ioutil.WriteFile("build", input, 0644)

	input, _ = ioutil.ReadFile(fmt.Sprintf("%s/bin/detect", shimDir))
	_ = ioutil.WriteFile("detect", input, 0644)

	input, _ = ioutil.ReadFile(fmt.Sprintf("%s/bin/release", shimDir))
	_ = ioutil.WriteFile("release", input, 0644)

	input, _ = ioutil.ReadFile(fmt.Sprintf("%s/bin/exports", shimDir))
	_ = ioutil.WriteFile("exports", input, 0644)

	fmt.Printf("at=descriptor file=%s api=%s id=%s version=%s name=%s stacks=%s\n\n",
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
	fmt.Printf("at=download file=%s url=%s\n\n", shimmedBuildpack, url)
	cmd := fmt.Sprintf(`curl --retry 3 --silent --location "%s" | tar xzm -C %s`, url, target_dir)

	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Printf("Failed to execute command: %s\n\n", cmd)
	}
	_ = os.Chdir(shimDir)

	defer fmt.Printf("at=cleanup file=%s\n\n", shimmedBuildpack)
	fstat, _ := file.Stat()
	cmd = fmt.Sprintf("tar cvfz %s %s", shimmedBuildpack, dir)

	_, err = exec.Command("bash", "-c", cmd).Output()
	defer os.Remove(shimmedBuildpack)

	if err != nil {
		fmt.Printf("Failed to execute command: %s\n\n", cmd)
		fmt.Println(os.Getwd())
	}

	fmt.Printf("at=send file=%s size=%d", shimmedBuildpack, fstat.Size())
	http.ServeFile(w, r, shimmedBuildpack)
	fmt.Printf("at=success file=%s", shimmedBuildpack)
}
