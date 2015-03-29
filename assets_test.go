package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestAssets_All(t *testing.T) {
	files := []string{
		"testdata/default/static/css/default.txt",
		"testdata/base.txt",
	}
	ass := NewAssets("assets", "asset_test.db")
	defer ass.Store.DeleteDatabase()

	// Save Assets
	for _, file := range files {
		_, err := ass.Save(file, "testdata")
		if err != nil {
			t.Error(err)
		}
	}

	// Retrieve saved assets
	for _, file := range files {
		f, err := ass.Get(strings.TrimPrefix(file, "testdata"))
		if err != nil {
			t.Error(err)
		} else if f.Name != filepath.Base(file) {
			t.Errorf("Expected %s got %s", filepath.Base(file), f.Name)
		}
	}

	// Remove assets
	for _, file := range files {
		err := ass.Remove(file)
		if err != nil {
			t.Error(err)
		}
		f, err := ass.Get(file)
		if err == nil {
			t.Errorf(" Expected nil, got %s", f.Name)
		}
	}

	// AddToStore
	ass.StaticDirs = []string{"testdata/default/static"}
	n := ass.AddToStore()
	if n != 2 {
		t.Errorf("Epected 2 got %d", n)
	}

	// sreve
	m := mux.NewRouter()
	m.HandleFunc("/{filename:.*}", ass.Serve).Methods("GET")

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/css/default.txt", nil)
	if err != nil {
		t.Error(err)
	}

	m.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Errorf("Expected 200 got %d", resp.Body)
	}
}
