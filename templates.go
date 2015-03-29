package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

type KTemplate struct {
	Store  *Store
	Bucket string
	Assets *Assets

	Cache map[string]*template.Template
}

type Config struct {
	Name      string `json:"name"`
	LayoutDir string `json:"layouts"`
	StaticDir string `json:"static_dir"`
	Version   string `json:"version"`
	Author    string `json:"author"`
	Repo      string `json:"repository"`
}

func (t *KTemplate) LoadToDB(pathname string) error {
	info, err := os.Stat(pathname)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("kesho KTemplate.LoadToDB: The pathname should be a valid directory")
	}
	if path.IsAbs(pathname) {
		return errors.New("kesho KTemplate.LOadToDB: The pathname should be relative")
	}

	configFile := filepath.Join(pathname, "config.json")

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	config := new(Config)

	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	layoutsPath := path.Join(pathname, config.LayoutDir)

	err = t.Store.CreateBucket(t.Bucket) // create bucket to hold all templates
	if err != nil {
		return err
	}
	err = filepath.Walk(layoutsPath, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		tdata, err := ioutil.ReadFile(root)
		if err != nil {
			return err
		}
		cleanPath := strings.TrimPrefix(strings.TrimPrefix(root, layoutsPath), "/")
		return t.Store.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(t.Bucket))
			if b == nil {
				return errors.New("kesho KTemplate.LoadToDB: No bucket for the templates found")
			}
			tb, err := b.CreateBucketIfNotExists([]byte(config.Name))
			if err != nil {
				return err
			}
			return tb.Put([]byte(cleanPath), tdata)
		})
	})
	if err != nil {
		return err
	}

	// Store the assets
	t.Assets.StaticDirs = []string{path.Join(pathname, config.StaticDir)}
	t.Assets.AddToStore()

	return nil

}
func (t *KTemplate) Render(w io.Writer, tmpl string, name string, data interface{}) error {
	render := t.Cache[tmpl]
	if render == nil {
		// missed  template
		if t.Exists(tmpl) {
			err := t.Load(tmpl)
			if err != nil {
				return err
			}
			return t.Render(w, tmpl, name, data)
		}
		return errors.New("kesho KTemplate.Render: No Template to render")
	}
	return render.ExecuteTemplate(w, name, data)
}

func (t *KTemplate) LoadFromDB() error {
	return t.Store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.Bucket))
		if b == nil {
			return errors.New("kesho KTemplate.LoadToDB: No bucket for the templates found")
		}

		return b.ForEach(func(k, v []byte) error {
			return t.loadTemplate(k, b)
		})
	})
}

func (t *KTemplate) Load(name string) error {
	return t.Store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.Bucket))
		if b == nil {
			return errors.New("kesho KTemplate.Load: No bucket for the templates found")
		}
		return t.loadTemplate([]byte(name), b)
	})
}
func (t *KTemplate) loadTemplate(name []byte, bucket *bolt.Bucket) error {
	var tmpl *template.Template
	b := bucket.Bucket(name)
	if b == nil {
		return errors.New("kesho KTemplate.LoadToDB: No bucket for the templates found")
	}

	// Adopted from the templates package on the ParseFiles implementation
	tmpl = template.New(string(name))
	err := b.ForEach(func(k, v []byte) error {
		var ntmpl *template.Template
		ntmpl = tmpl.New(string(k))
		_, terr := ntmpl.Parse(string(v))
		return terr
	})
	if err != nil {
		return err
	}

	//Add to cache
	t.Cache[tmpl.Name()] = tmpl
	return nil
}
func (t *KTemplate) Exists(name string) bool {
	err := t.Store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.Bucket))
		if b == nil {
			return errors.New("No bucket buddy")
		}
		nb := b.Bucket([]byte(name))
		if nb == nil {
			return errors.New("No bucket buddy")

		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}
