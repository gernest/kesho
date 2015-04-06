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
		Convey("Storing Assets", func() {
			Convey("With a given path prefix", func() {
				n := "web/config.json"
				f, err := ass.Save(n, "web")

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("Without path prefix", func() {
				n := "web/config.json"
				f, err := ass.Save(n)

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("With a file that does not exist", func() {
				n := "web/bogus.json"
				f, err := ass.Save(n, "web")

				So(err, ShouldNotBeNil)
				So(f, ShouldBeNil)
			})
		})

		Convey("Retrieving assets", func() {

			Convey("With a valid file", func() {
				n := "web/config.json"
				ass.Save(n, "web")

				f, err := ass.Get("config.json")

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("With a file that does not exist", func() {
				f, err := ass.Get("kemi.json")

				So(err, ShouldNotBeNil)
				So(f, ShouldBeNil)
			})
			Convey("With absolute path", func() {
				n := "web/config.json"
				ass.Save(n)
				f, err := ass.Get(n)

				So(err, ShouldBeNil)
				So(f.Name, ShouldEqual, "config.json")
			})
			Convey("With absolute path and a file that does not exist", func() {
				f, err := ass.Get("/kemi.json")

				So(err, ShouldNotBeNil)
				So(f, ShouldBeNil)
			})

		})
		Convey("Deleting Asset", func() {
			n := "web/config.json"
			ass.Save(n, "web")

			err := ass.Delete("config.json")

			f, gerr := ass.Get("config.json")

			So(err, ShouldBeNil)
			So(gerr, ShouldNotBeNil)
			So(f, ShouldBeNil)
		})
		Convey("Load Assets from directories", func() {
			ass.LoadDirs(dirs...)
			file, err := ass.Get("css/docs.css")

			So(err, ShouldBeNil)
			So(file.Name, ShouldEqual, "docs.css")
		})
		Convey("Serving assets", func() {
			m := mux.NewRouter()
			m.HandleFunc("/static/{filename:.*}", ass.Serve).Methods("GET")
			w := httptest.NewRecorder()
			Convey("A file that exists", func() {
				r, _ := http.NewRequest("GET", "/static/css/docs.css", nil)

				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("A file does not exist", func() {
				r, _ := http.NewRequest("GET", "/static/css/horses.css", nil)

				m.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}
