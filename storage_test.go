package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestStorage(t *testing.T) {
	tstore := NewStorage("storage_test.db", 0600)
	bList := []string{"bucket", "bucket", "bucket"}

	defer tstore.DeleteDatabase()

	Convey("Working with boltdb store", t, func() {
		Convey("Creating New Record", func() {
			n := tstore.CreateDataRecord("base", "record", []byte("data"), bList...)
			So(n.Error, ShouldBeNil)
			So(n.Data, ShouldNotBeNil)
			So(string(n.Data), ShouldEqual, "data")
		})

		Convey("Getting records from database", func() {
			n := tstore.CreateDataRecord("base", "record2", []byte("data"), bList...)
			g := n.GetDataRecord("base", "record2", "bucket", "bucket", "bucket")
			So(n.Error, ShouldBeNil)
			So(g.Error, ShouldBeNil)
			So(string(n.Data), ShouldEqual, string((g.Data)))
		})

		Convey("Updating database Record", func() {
			n := tstore.CreateDataRecord("base", "record2", []byte("data"), bList...)
			up := n.UpdateDataRecord("base", "record2", []byte("data update"), bList...)

			uprec := up.GetDataRecord("base", "record2", bList...)

			So(up.Error, ShouldBeNil)
			So(uprec.Error, ShouldBeNil)
			So(string(uprec.Data), ShouldEqual, "data update")
		})

		Convey("Get All key pairs from a given bucket", func() {
			dd := []struct {
				key, value string
			}{
				{"moja", "moja"},
				{"mbili", "mbili"},
				{"tatu", "tatu"},
				{"nne", "nne"},
			}
			nest := [][]string{
				[]string{"a", "b"},
				[]string{"c", "d"},
				[]string{"e", "d"},
				[]string{"g", "h"},
			}

			buck := "bucky"

			for k, v := range dd {
				tstore.CreateDataRecord(buck, v.key, []byte(v.value), nest[k]...)
			}
			all := tstore.GetAll(buck)
			allnest := all.GetAll(buck, nest[0]...)

			So(len(all.DataList), ShouldEqual, 4)
			So(len(allnest.DataList), ShouldEqual, 1)
			So(string(allnest.DataList[dd[0].key]), ShouldEqual, dd[0].value)
		})

		Convey("Remove a record from the database", func() {
			tstore.CreateDataRecord("base", "record2", []byte("data"), "bucket", "bucket", "bucket")
			tstore.RemoveDataRecord("base", "record2", "bucket", "bucket", "bucket")

			g := tstore.GetDataRecord("base", "record2", "bucket", "bucket", "bucket")

			So(g.Data, ShouldBeNil)
		})
	})
}
