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
	"errors"
	"github.com/gorilla/mux"
	"sync"
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
	Bucket string
	Store  Storage
}

func NewAssets(bucket, storeName string) *Assets {
	return &Assets{
		Bucket: bucket,
		Store:  NewStorage(storeName, 0600),
	}
}

func (ass *Assets) Save(filename string, prefix ...string) (file *File, err error) {
	var pref string
	pref = ""
	if len(prefix) > 0 {
		pref = prefix[0]
	}
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
	key := strings.TrimPrefix(filename, pref)
	if filepath.IsAbs(key) {
		key = strings.TrimPrefix(key, "/")
	}
	s := ass.Store.CreateDataRecord(ass.Bucket, key, packedData)
	if s.Error != nil {
		return nil, s.Error
	}
	return file, err
}

func (ass *Assets) LoadDirs(dirs ...string) {
	for _, dir := range dirs {
		ass.loadDir(dir)
	}
}

func (ass *Assets) loadDir(dir string) {
	wg := new(sync.WaitGroup)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		wg.Add(1)
		go func(file, dir string, wg *sync.WaitGroup) {
			_, serr := ass.Save(file, dir)
			if serr != nil {
				log.Println(err)
			}
			log.Println(file, "--loaded")
			wg.Done()
		}(path, dir, wg)
		return nil
	})
	wg.Wait()
}

func (ass *Assets) Serve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	file, err := ass.Get(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	ass.ServeContent(w, r, file)
}

func (ass *Assets) ServeContent(w http.ResponseWriter, r *http.Request, file *File) {
	reader := bytes.NewReader([]byte(file.Body))
	http.ServeContent(w, r, file.Name, file.UpdatedAt, reader)
}

func (ass *Assets) Delete(key string) error {
	rm := ass.Store.RemoveDataRecord(ass.Bucket, key)
	return rm.Error
}

func (ass *Assets) Get(key string) (*File, error) {
	g := ass.Store.GetDataRecord(ass.Bucket, key)
	if g.Error != nil {
		return nil, g.Error
	}
	if g.Data == nil && filepath.IsAbs(key) {
		g = ass.Store.GetDataRecord(ass.Bucket, strings.TrimPrefix(key, "/"))
		if g.Error != nil {
			return nil, g.Error
		}
	} else if g.Data == nil {
		return nil, errors.New("kesho.Assets: The requested file was deleted")
	}

	file := new(File)
	err := json.Unmarshal(g.Data, file)
	if err != nil {
		return nil, err
	}
	return file, nil
}
