package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/logger"
	"github.com/heroku/cnb-shim"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage:", os.Args[0], "TARGET_BUILDPACK_DIR", "LAYERS_DIR", "PLATFORM_DIR")
		return
	}

	log, err := logger.DefaultLogger(os.Args[3])
	if err != nil {
		log.Info(err.Error())
		os.Exit(1)
	}

	targetDir := os.Args[1]
	layersDir := os.Args[2]

	appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Info(err.Error())
		os.Exit(2)
	}

	err = releaser.WriteLaunchMetadata(appDir, layersDir, targetDir, log)
	if err != nil {
		log.Info(err.Error())
		os.Exit(3)
	}
}
