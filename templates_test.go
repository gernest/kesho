package main

//
//import (
//	"bytes"
//	"html/template"
//	"testing"
//)
//
//func TestKTemplate_All(t *testing.T) {
//	ass := NewAssets("assets", "asset_test.db")
//	defer ass.Store.DeleteDatabase()
//
//	templatePath := []string{"testdata/default", "web"}
//
//	ktemp := &KTemplate{
//		Bucket: "templates",
//		Store:  NewStore("asset_test.db",0600,nil),
//		Assets: ass,
//		Cache:  make(map[string]*template.Template),
//	}
//
//	// Load templates from disc and save to database
//
//	for _, v := range templatePath {
//		err := ktemp.LoadToDB(v)
//		if err != nil {
//			t.Error(err)
//		}
//	}
//
//	// Load templates from database into memory.
//	err := ktemp.LoadFromDB()
//	if err != nil {
//		t.Error(err)
//	}
//
//	tmplNames := []string{"default"}
//	for _, v := range tmplNames {
//		if !ktemp.Exists(v) {
//			t.Error(ktemp.Cache)
//		}
//	}
//
//	parsedTemp := ktemp.Cache["default"]
//
//	if parsedTemp.Name() != "default" {
//		t.Errorf("Expected default got %s", parsedTemp.Name())
//	}
//
//	// Render
//	out := new(bytes.Buffer)
//	err = ktemp.Render(out, "default", "index.txt", ktemp)
//	if err != nil {
//		t.Error(err)
//	}
//
//	// Remove the default template
//	out.Reset()
//	delete(ktemp.Cache, "default")
//
//	// make sure it aint there
//	if ktemp.Cache["default"] != nil {
//		t.Errorf("Expected nil got %s", ktemp.Cache["default"].Name())
//	}
//
//	// Lets hope it reloads the template
//	err = ktemp.Render(out, "default", "post/post.txt", ktemp)
//	if err != nil {
//		t.Error(err)
//	}
//
//	// Should be loaded by now
//	if ktemp.Cache["default"] == nil {
//		t.Errorf("Fuck IISI AGAIN", out.String())
//	}
//}
