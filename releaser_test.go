package releaser_test

import (
	"path/filepath"
	"testing"

	"github.com/heroku/cnb-shim"
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
