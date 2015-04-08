package main

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestKTemplates(t *testing.T) {
	ass := NewAssets("myass", "templ_test.db")
	defer ass.Store.DeleteDatabase()
	ktemp := &KTemplate{Bucket: "my templs", Store: NewStorage("templ_test.db", 0660), Assets: ass}

	Convey("Testing templates", t, func() {
		Convey("Loading a directory of templates to database", func() {
			Convey("Given A valid Path", func() {
				dir := "web"
				err := ktemp.LoadToDB(dir)
				So(err, ShouldBeNil)
			})
			Convey("Given a wrong path", func() {
				dir := "shit"
				err := ktemp.LoadToDB(dir)
				So(err, ShouldNotBeNil)
			})
			Convey("WHen given a filename instead of a directory", func() {
				dir := "web/config.json"
				err := ktemp.LoadToDB(dir)
				So(err, ShouldNotBeNil)
			})

		})
		Convey("Loading single template from database", func() {
			err := ktemp.LoadSingle("web")

			So(err, ShouldBeNil)
		})
		Convey("Loda templates from database", func() {
			err := ktemp.LoadEm()

			So(err, ShouldBeNil)
			So(len(ktemp.Cache), ShouldNotEqual, 0)
			So(ktemp.Exists("web"), ShouldBeTrue)
			So(len(ktemp.AuthTempl), ShouldEqual, 4)
		})
		Convey("Rendering templates", func() {
			Convey("When the template is loaded", func() {
				buf := new(bytes.Buffer)
				data := make(map[string]interface{})
				data["Title"] = "wa web"
				err := ktemp.Render(buf, "web", "accounts/index.html", data)

				So(err, ShouldBeNil)
				So(buf.String(), ShouldContainSubstring, "wa web")
			})
			Convey("When the template is not loaded", func() {
				buf := new(bytes.Buffer)
				data := make(map[string]interface{})
				data["Title"] = "wa web"
				err := ktemp.Render(buf, "keshotena", "accounts/index.html", data)

				So(err, ShouldNotBeNil)
			})

		})
	})
}
