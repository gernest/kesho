package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type File struct {
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"create_at"`
	UpdatedAt time.Time `json:"update_at"`
}

type Assets struct {
	Bucket     string
	Store      *Store
	StaticDirs []string
}

func NewAssets(bucket, storeName string) *Assets {
	return &Assets{
		Bucket: bucket,
		Store:  NewStore(storeName, 0600, nil),
	}
}

func (ass *Assets) Save(filename, prefix string) (file *File, err error) {
	if _, err = os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	file = new(File)
	file.Name = filepath.Base(filename)
	file.Body = string(data)

	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
	packedData, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	key := strings.TrimPrefix(filename, prefix)
	if filepath.IsAbs(key) {
		key = strings.TrimPrefix(key, "/")
	}
	if ass.Store.CreateRecord(ass.Bucket, key, packedData).Error != nil {
		log.Println(err)
	}
	return file, err
}

// AddToSTore save the files in the StaticDirs to database and returs the number
// of files saved
func (ass *Assets) AddToStore() (n int) {
	for _, dir := range ass.StaticDirs {
		n += ass.loadDir(dir)
	}
	return
}

func (ass *Assets) loadDir(dir string) (n int) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		} else {
			_, serr := ass.Save(path, dir)
			if serr != nil {
				return serr
			}
			n += 1
		}
		return nil
	})
	return
}

func (ass *Assets) Serve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	file, err := ass.Get(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ass.ServeContent(w, r, file)
}

// ServeContent Borrows heavily on `http.ServeContent`
func (ass *Assets) ServeContent(w http.ResponseWriter, r *http.Request, file *File) {

	// Checks for modified time header, this code is borrowed from the standard library
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && file.UpdatedAt.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Last-Modified", file.UpdatedAt.UTC().Format(http.TimeFormat))

	// TODO: checks for Etag, and heck I have no idea what this thing is but i guess it is important.

	code := http.StatusOK

	ctypes, haveType := w.Header()["Content-Type"]
	var ctype string
	if !haveType {
		ctype = mime.TypeByExtension(filepath.Ext(file.Name))
		w.Header().Set("Content-Type", ctype)
	} else if len(ctypes) > 0 {
		ctype = ctypes[0]
	}
	w.WriteHeader(code)
	if r.Method != "HEAD" {
		w.Write([]byte(file.Body))
	}
}

func (ass *Assets) Remove(key string) error {
	return ass.Store.removeRecord(ass.Bucket, key).Error
}

func (ass *Assets) Get(key string) (*File, error) {
	ass.Store.GetRecord(ass.Bucket, key)
	if ass.Store.Error != nil {
		return nil, ass.Store.Error
	}
	if ass.Store.Data == nil && filepath.IsAbs(key) {
		ass.Store.GetRecord(ass.Bucket, strings.TrimPrefix(key, "/"))
		if ass.Store.Error != nil {
			return nil, ass.Store.Error
		}
	}
	file := new(File)
	err := json.Unmarshal(ass.Store.Data, file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Just Incase I want to support Range requests in the future
//----------------------------------------------------------
//-  Range Request and Etags stuffs  adopted from net/http
//----------------------------------------------------------
//func checkETag(w http.ResponseWriter,r *http.Request)(rangeReq string,done bool){
//	etag:=w.Header().Get("Etag")
//	rangeReq=getHeader("Range",r)
//
//	if ir:=getHeader("If-Range",r);ir!=""&&ir!=etag {
//		rangeReq=""
//	}
//
//	if inm:=getHeader("If-None-Match",r);inm!="" {
//		if etag=="" {
//			return rangeReq,false
//		}
//		if r.Method!="GET"&&r.Method!="HEAD" {
//			return rangeReq,false
//		}
//		if inm == etag || inm == "*" {
//			h := w.Header()
//			delete(h, "Content-Type")
//			delete(h, "Content-Length")
//			w.WriteHeader(http.StatusNotModified)
//			return "", true
//		}
//	}
//	return rangeReq,false
//}
//
//func getHeader(key string, r *http.Request) string{
//	if v:=r.Header[key];len(v)>0 {
//		return  v[0]
//	}
//	return ""
//}