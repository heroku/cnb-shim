package cnbshim_test


import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/logger"
	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/layers"
	releaser "github.com/heroku/cnb-shim"
)

func TestReadProcfileWithWebAndWorker(t *testing.T) {
	appWithProcfile := filepath.Join("test", "fixtures", "app_with_procfile")
	got, err := releaser.ReadProcfile(appWithProcfile)
	if err != nil {
		t.Error(err.Error())
	}

	if got["web"] != "node index.js" {
		t.Errorf("Expected 'web' process type of 'node index.js'; got %s", got)
	}

	if got["worker"] != "node worker.js" {
		t.Errorf("Expected 'web' process type of 'node worker.js'; got %s", got)
	}
}

func TestReadProcfileWithoutProcfile(t *testing.T) {
	appWithProcfile := filepath.Join("test", "fixtures", "app_without_procfile")
	got, err := releaser.ReadProcfile(appWithProcfile)
	if err != nil {
		t.Error(err.Error())
	}

	if len(got) != 0 {
		t.Errorf("Expected no process types; got %s", got)
	}
}

func TestReadProcfileWithEmptyProcfile(t *testing.T) {
	appWithProcfile := filepath.Join("test", "fixtures", "app_with_empty_procfile")
	got, err := releaser.ReadProcfile(appWithProcfile)
	if err != nil {
		t.Error(err.Error())
	}

	if len(got) != 0 {
		t.Errorf("Expected no process types; got %s", got)
	}
}

func TestExecReleaseWithoutDefaultProcs(t *testing.T) {
	buildpack := filepath.Join("test", "fixtures", "buildpack_without_default_procs")
	app := filepath.Join("test", "fixtures", "app_with_empty_procfile")
	got, err := releaser.ExecReleaseScript(app, buildpack)
	if err != nil {
		t.Error(err.Error())
	}

	if len(got.DefaultProcessTypes) != 0 {
		t.Errorf("Expected no process types; got %s", got)
	}
}

func TestExecReleaseWithDefaultProcs(t *testing.T) {
	buildpack := filepath.Join("test", "fixtures", "buildpack_with_default_procs")
	app := filepath.Join("test", "fixtures", "app_with_empty_procfile")
	got, err := releaser.ExecReleaseScript(app, buildpack)
	if err != nil {
		t.Error(err.Error())
	}

	expected := "java -jar myapp.jar"
	if got.DefaultProcessTypes["web"] != expected {
		t.Errorf("Expected 'web' process type of '%s'; got %s", expected, got)
	}
}

func TestWriteLaunchMetadata(t *testing.T) {
	buildpack := filepath.Join("test", "fixtures", "buildpack_with_default_procs")
	app := filepath.Join("test", "fixtures", "app_with_procfile")
	layersDir, err := ioutil.TempDir("", "layers")
	if err != nil {
		t.Error(err.Error())
	}

	log, err := logger.DefaultLogger(os.TempDir())
	if err != nil {
		t.Error(err.Error())
	}

	err = releaser.WriteLaunchMetadata(app, layersDir, buildpack, log)
	if err != nil {
		t.Error(err.Error())
	}

	l := layers.Metadata{}

	_, err = toml.DecodeFile(filepath.Join(layersDir, "launch.toml"), &l)
	if err != nil {
		t.Error(err.Error())
	}

	if len(l.Processes) != 2 {
		t.Errorf("Expected 2 process type; got %d", len(l.Processes))
	}

	foundWeb := false
	foundWorker := false
	for _, p := range l.Processes {
		if p.Type == "web" {
			foundWeb = true
			expected := "node index.js"
			if p.Command != expected {
				t.Errorf("Expected 'web' process type of '%s'; got %s", expected, p.Command)
			}
		} else if p.Type == "worker" {
			foundWorker = true
			expected := "node worker.js"
			if p.Command != expected {
				t.Errorf("Expected 'worker' process type of '%s'; got %s", expected, p.Command)
			}
		}
	}

	if !foundWeb {
		t.Errorf("Expected 'web' process type; got %s", l.Processes)
	}

	if !foundWorker {
		t.Errorf("Expected 'worker' process type; got %s", l.Processes)
	}
}

