package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_downloadBuildPack(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this is a buildpack"))
	})

	server := httptest.NewServer(handler)
	defer server.Close()
	file, err := downloadBuildpack(server.URL)
	defer os.Remove(file)

	if err != nil {
		t.Fatal("error creating buildpack file")
	}

	contents, _ := ioutil.ReadFile(file)

	if string(contents) != "this is a buildpack" {
		t.Fatal("incorrect buildpack contents")
	}
}
