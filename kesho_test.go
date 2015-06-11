package main

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestKesh(t *testing.T) {
	cfg := new(KConfig)
	cfg.SessDB = "sess_test.db"
	cfg.MainDB = "main_test.db"
	k := NewKesho(cfg)

	if err := k.Templ.LoadToDB("web"); err != nil {
		t.Error(err)
	}
	kser := k.TestServer()
	uri := kser.URL

	defer k.Store.DeleteDatabase()
	defer os.Remove(cfg.SessDB)
	defer kser.Close()

	Convey("Handlers", t, func() {
		Convey("Home Handler", func() {
			res, err := http.Get(uri)
			buf := new(bytes.Buffer)
			io.Copy(buf, res.Body)

			So(err, ShouldBeNil)
			So(res.StatusCode, ShouldEqual, 200)
			So(buf.String(), ShouldContainSubstring, " Blogu yako")
		})
	})

}
