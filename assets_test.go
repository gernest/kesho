package main

import (
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssets(t *testing.T) {
	ass := NewAssets("assets", "asset_test.db")
	defer ass.Store.DeleteDatabase()
	dirs := []string{"web/static"}

	Convey("Testing Assets", t, func() {
		Convey("Store Assets", func() {
			Convey("With prefix", func() {
				n := "web/config.json"
				f, err := ass.Save(n, "web")

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("Without Prefix", func() {
				n := "web/config.json"
				f, err := ass.Save(n)

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("With bogus file", func() {
				n := "web/bogus.json"
				f, err := ass.Save(n, "web")

				So(err, ShouldNotBeNil)
				So(f, ShouldBeNil)
			})
		})

		Convey("Retrieve some assets", func() {

			Convey("With a valid file", func() {
				n := "web/config.json"
				ass.Save(n, "web")

				f, err := ass.Get("config.json")

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("With a file not in stored", func() {
				f, err := ass.Get("kemi.json")

				So(err, ShouldNotBeNil)
				So(f, ShouldBeNil)
			})
			Convey("With absolute path", func() {
				f, err := ass.Get("/config.json")

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("With absolute path and wrong file", func() {
				f, err := ass.Get("/kemi.json")

				So(err, ShouldNotBeNil)
				So(f, ShouldBeNil)
			})

		})
		Convey("Delet Asset", func() {
			n := "web/config.json"
			ass.Save(n, "web")

			err := ass.Delete("config.json")

			f, gerr := ass.Get("config.json")

			So(err, ShouldBeNil)
			So(gerr, ShouldNotBeNil)
			So(f, ShouldBeNil)
		})
		Convey("Load Assets", func() {
			ass.LoadDirs(dirs...)
			file, err := ass.Get("css/docs.css")

			So(err, ShouldBeNil)
			So(file.Name, ShouldEqual, "docs.css")
		})
		Convey("Serving assets", func() {
			m := mux.NewRouter()
			m.HandleFunc("/static/{filename:.*}", ass.Serve).Methods("GET")
			w := httptest.NewRecorder()
			Convey("Present files", func() {
				r, _ := http.NewRequest("GET", "/static/css/docs.css", nil)

				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("File not current stores", func() {
				r, _ := http.NewRequest("GET", "/static/css/horses.css", nil)

				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}
