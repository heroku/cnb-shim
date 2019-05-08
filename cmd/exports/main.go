package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/buildpack/libbuildpack/logger"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "EXPORTS_PATH", "PLATFORM_DIR")
		return
	}

	log, err := logger.DefaultLogger(os.Args[3])
	if err != nil {
		log.Info(err.Error())
		os.Exit(1)
	}

	exportsPath := os.Args[1]

	data, err := ioutil.ReadFile(exportsPath)
	if err != nil {
		log.Info(err.Error())
		os.Exit(2)
	}

	contents := string(data)
	lines := strings.SplitN(contents, "\n", 2)

	for _, line := range lines {
		components := strings.Split(line, "=")
		ioutil.WriteFile(components[0], []byte(components[1]), 0644)
	}
}
