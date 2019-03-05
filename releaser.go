package releaser

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
)

type Release struct {
	DefaultProcessTypes map[string]string `yaml:"default_process_types,omitempty"`
}

func WriteLaunchMetadata(appDir, layersDir, targetBuildpackDir string, log logger.Logger) error {
	release, err := ExecReleaseScript(appDir, targetBuildpackDir)
	if err != nil {
		return err
	}

	procfile, err := ReadProcfile(appDir)
	if err != nil {
		return err
	}

	processTypes := make(map[string]string)
	for name, command := range release.DefaultProcessTypes {
		processTypes[name] = command
	}

	for name, command := range procfile {
		processTypes[name] = command
	}

	processes := layers.Processes{}
	for name, command := range processTypes {
		processes = append(processes, layers.Process{
			Type:    name,
			Command: command,
		})
	}

	l := layers.NewLayers(layersDir, log)

	return l.WriteApplicationMetadata(layers.Metadata{
		Processes: processes,
	})
}

func ExecReleaseScript(appDir, targetBuildpackDir string) (Release, error) {
	releaseScript := filepath.Join(targetBuildpackDir, "bin", "release")
	_, err := os.Stat(releaseScript)
	if !os.IsNotExist(err) {
		cmd := exec.Command(releaseScript, appDir)
		cmd.Env = os.Environ()

		out, err := cmd.Output()
		if err != nil {
			return Release{DefaultProcessTypes: make(map[string]string)}, err
		}

		release := Release{}

		return release, yaml.Unmarshal(out, &release)
	} else {
		return Release{DefaultProcessTypes: make(map[string]string)}, nil
	}

}

func ReadProcfile(appDir string) (map[string]string, error) {
	processTypes := make(map[string]string)
	procfile := filepath.Join(appDir, "Procfile")
	_, err := os.Stat(procfile)
	if !os.IsNotExist(err) {

		procfileText, err := ioutil.ReadFile(procfile)
		if err != nil {
			return processTypes, err
		}

		return processTypes, yaml.Unmarshal(procfileText, &processTypes)
	} else {
		return processTypes, nil
	}
}