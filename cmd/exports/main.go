package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack/logger"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "EXPORTS_PATH", "PLATFORM_DIR")
		return
	}

	exportsPath := os.Args[1]
	platformDir := os.Args[2]

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

	contents := string(data)
	lines := strings.SplitN(contents, "\n", 2)

	for _, line := range lines {
		fmt.Println(line)
		components := strings.Split(line, "=")
		export := strings.Split(components[0], " ")
		if strings.TrimSpace(export[0]) == "export" {
			fmt.Println("Export 1:")
			fmt.Println(export[1])
			file := filepath.Join(platformDir, strings.TrimSpace(export[1]))
			err := ioutil.WriteFile(file, []byte(components[1]), 0644)
			if err != nil {
				log.Info(err.Error())
				os.Exit(3)
			}
		}
	}
}
