package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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

			Convey("Records in nested buckets", func() {
				n := tstore.CreateDataRecord("base", "record2", []byte("data"), bList...)
				So(n.Error, ShouldBeNil)

				Convey("Record Found", func() {
					g := n.GetDataRecord("base", "record2", bList...)
					So(g.Error, ShouldBeNil)
					So(string(n.Data), ShouldEqual, string((g.Data)))
				})
				Convey("Record not found", func() {
					g := n.GetDataRecord("base", "recordz", bList...)
					So(g.Error, ShouldNotBeNil)
					So(g.Data, ShouldBeNil)
				})
				Convey("With a wrong bucket", func() {
					g := n.GetDataRecord("base2", "record2", bList...)
					So(g.Error, ShouldNotBeNil)
					So(g.Data, ShouldBeNil)
				})
				Convey("Wrong bucket list", func() {
					list1 := []string{"bucket", "bucket", "chahchacha"}
					list2 := []string{"bucket", "chachacha", "bucket"}
					list3 := []string{"chachacha", "bucket", "bucket"}

					g1 := n.GetDataRecord("base", "record2", list1...)
					g2 := n.GetDataRecord("base", "record2", list2...)
					g3 := n.GetDataRecord("base", "record2", list3...)

					So(g1.Error, ShouldNotBeNil)
					So(g1.Data, ShouldBeNil)

					So(g2.Error, ShouldNotBeNil)
					So(g2.Data, ShouldBeNil)

					So(g3.Error, ShouldNotBeNil)
					So(g3.Data, ShouldBeNil)
				})
			})
			Convey("Records not in a nested bucket", func() {
				n := tstore.CreateDataRecord("base", "record2", []byte("data"))
				So(n.Error, ShouldBeNil)

				Convey("Record found", func() {
					g := n.GetDataRecord("base", "record2")
					So(g.Error, ShouldBeNil)
					So(string(n.Data), ShouldEqual, string((g.Data)))
				})

				Convey("Wrong bucket list", func() {
					g := n.GetDataRecord("base", "record2", "bug")
					So(g.Error, ShouldNotBeNil)
					So(g.Data, ShouldBeNil)
				})
				Convey("With  bucket name ot in the database", func() {
					g := n.GetDataRecord("base2", "record2", "bug")
					So(g.Error, ShouldNotBeNil)
					So(g.Data, ShouldBeNil)
				})

			})

		})

		Convey("Updating database Record", func() {
			Convey("With nested buckets", func() {
				n := tstore.CreateDataRecord("base", "record2", []byte("data"), bList...)

				Convey("Record Exist", func() {
					up := n.UpdateDataRecord("base", "record2", []byte("data update"), bList...)

					uprec := up.GetDataRecord("base", "record2", bList...)

					So(up.Error, ShouldBeNil)
					So(uprec.Error, ShouldBeNil)
					So(string(uprec.Data), ShouldEqual, "data update")
				})
				Convey("Record does not exist", func() {
					up := n.UpdateDataRecord("base", "recordnp", []byte("data update"), bList...)

					So(up.Error, ShouldNotBeNil)
					So(up.Data, ShouldBeNil)
				})
				Convey("Wrong bucket", func() {
					up := n.UpdateDataRecord("basenp", "record2", []byte("data update"), bList...)

					So(up.Error, ShouldNotBeNil)
					So(up.Data, ShouldBeNil)
				})
				Convey("Wrong Bucket list", func() {
					list := []string{"bucket", "bucket", "chachacha"}
					up := n.UpdateDataRecord("base", "record2", []byte("data update"), list...)

					So(up.Error, ShouldNotBeNil)
					So(up.Data, ShouldBeNil)
				})
			})
			Convey("Without nested buckets", func() {
				n := tstore.CreateDataRecord("basenot", "not_used", []byte("data"))
				Convey("Record Exist", func() {
					up := n.UpdateDataRecord("basenot", "not_used", []byte("data update"))

					uprec := up.GetDataRecord("basenot", "not_used")

					So(up.Error, ShouldBeNil)
					So(uprec.Error, ShouldBeNil)
					So(string(uprec.Data), ShouldEqual, "data update")
				})
				Convey("Record does not exist", func() {
					up := n.UpdateDataRecord("basenot", "nat_used2", []byte("data update"))

					So(up.Error, ShouldNotBeNil)
					So(up.Data, ShouldBeNil)
				})
				Convey("Wrong bucket", func() {
					up := n.UpdateDataRecord("basenot", "nat_used", []byte("data update"))

					So(up.Error, ShouldNotBeNil)
					So(up.Data, ShouldBeNil)
				})
			})

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
			tstore.CreateDataRecord("base", "record2", []byte("data"), bList...)
			tstore.RemoveDataRecord("base", "record2", bList...)

			g := tstore.GetDataRecord("base", "record2", bList...)

			So(g.Data, ShouldBeNil)
		})
	})
}
