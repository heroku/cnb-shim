package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/buildpack/libbuildpack/logger"
	"github.com/heroku/cnb-shim"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage:", os.Args[0], "EXPORTS_PATH", "PLATFORM_DIR", "ENV_DIR")
		return
	}

	exportsPath := os.Args[1]
	platformDir := os.Args[2]
	envDir := os.Args[3]

	log, err := logger.DefaultLogger(platformDir)
	if err != nil {
		log.Info(err.Error())
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(exportsPath)
	if err != nil {
		log.Info(err.Error())
		os.Exit(2)
	}

	if err := cnbshim.DumpExportsFile(string(data), envDir); err != nil {
		log.Info(err.Error())
		os.Exit(3)
	}
}
