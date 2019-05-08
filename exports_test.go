package cnbshim_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heroku/cnb-shim"
)

func TestWriteEnvFile(t *testing.T) {
	envDir, err := ioutil.TempDir("", "env")
	if err != nil {
		t.Error(err.Error())
	}

	expectedFilename := "FOO"
	expected := "bar"
	err = cnbshim.WriteEnvFile(envDir, expectedFilename, expected)
	if err != nil {
		t.Error(err.Error())
	}

	expectedFile := filepath.Join(envDir, expectedFilename)
	actual, err := ioutil.ReadFile(expectedFile)
	if err != nil {
		t.Error(err.Error())
	}

	if string(actual) != expected {
		t.Errorf("Expected %s; got %s", expected, actual)
	}
}

func TestDumpExportsFile(t *testing.T) {
	envDir, err := ioutil.TempDir("", "env")

	exports := `
export FOO=bar
`
	err = cnbshim.DumpExportsFile(exports, envDir)
	if err != nil {
		t.Error(err.Error())
	}

	expectedFile := filepath.Join(envDir, "FOO")
	expected := "bar"
	actual, err := ioutil.ReadFile(expectedFile)
	if err != nil {
		t.Error(err.Error())
	}

	if strings.TrimSpace(string(actual)) != expected {
		t.Errorf("Expected %s; got %s", expected, actual)
	}
}