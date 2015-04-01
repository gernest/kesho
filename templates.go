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
	"github.com/gernest/authboss"

	"regexp"
)

type KTemplate struct {
	Store  *Store
	Bucket string
	Assets *Assets

	AuthTempl map[string]*template.Template
	Cache     map[string]*template.Template
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
		return t.Store.createRecord(t.Bucket, cleanPath, tdata, config.Name).Error
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
	var layout *template.Template
	var authTempl map[string][]byte
	authTempl = make(map[string][]byte)

	var funcMap = template.FuncMap{
		"title": strings.Title,
		"mountpathed": func(location string) string {
			if authboss.Cfg.MountPath == "/" {
				return location
			}
			return path.Join(authboss.Cfg.MountPath, location)
		},
	}

	b := bucket.Bucket(name)
	if b == nil {
		return errors.New("kesho KTemplate.LoadToDB: No bucket for the templates found")
	}

	// Adopted from the templates package on the ParseFiles implementation
	tmpl = template.New(string(name))

	err := b.ForEach(func(k, v []byte) error {
		var ntmpl *template.Template
		ntmpl = tmpl.New(string(k))
		re := regexp.MustCompile("^*[.]html.tpl$")
		if re.Match(k) {
			authTempl[string(k)] = v
			return nil
		}
		_, terr := ntmpl.Parse(string(v))
		return terr
	})
	if err != nil {
		return err
	}
	if len(authTempl) > 0 {
		layout = template.Must(template.New("layout").Funcs(funcs).Parse(string(authTempl["layout.html.tpl"])))
		t.AuthTempl = make(map[string]*template.Template)
		for k, v := range authTempl {
			if k == "layout.html.tpl" {
				continue
			}
			clone, err := layout.Clone()
			if err != nil {
				panic(err)
			}

			_, err = clone.New("authboss").Funcs(funcMap).Parse(string(v))
			if err != nil {
				panic(err)
			}
			t.AuthTempl[k] = clone
		}
		t.Cache[layout.Name()] = layout
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
