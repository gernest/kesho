package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bytes"
	"github.com/gorilla/mux"
)

type File struct {
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"create_at"`
	UpdatedAt time.Time `json:"update_at"`
}

func (f *File) Size() int64 {
	return int64(len(f.Body))
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
	reader := bytes.NewReader([]byte(file.Body))
	http.ServeContent(w, r, file.Name, file.UpdatedAt, reader)
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
