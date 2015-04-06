package main

import (
	. "github.com/smartystreets/goconvey/convey"
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
			n := "web/config.json"
			ass.Save(n, "web")

			f, err := ass.Get("config.json")

			So(err, ShouldBeNil)
			So(f.Name, ShouldEqual, "config.json")

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
	})
}
