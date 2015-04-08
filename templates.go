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

	"github.com/gernest/authboss"
	"regexp"
)

type KTemplate struct {
	Store  Storage
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
		s := t.Store.CreateDataRecord(t.Bucket, cleanPath, tdata, config.Name)
		return s.Error
	})
	if err != nil {
		return err
	}
	// Store the assets
	staticDirs := []string{path.Join(pathname, config.StaticDir)}
	t.Assets.LoadDirs(staticDirs...)
	return nil

}
func (t *KTemplate) Render(w io.Writer, tmpl string, name string, data interface{}) error {
	render := t.Cache[tmpl]
	if render == nil {
		if err := t.LoadSingle(tmpl); err == nil {
			return t.Render(w, tmpl, name, data)
		}
		return errors.New("kesho KTemplate.Render: No Template to render")
	}
	return render.ExecuteTemplate(w, name, data)
}

func (t *KTemplate) LoadEm() error {
	b := t.Store.GetAll(t.Bucket)
	if b.Error != nil {
		return b.Error
	}
	if len(b.DataList) == 0 {
		return errors.New("No templates in the database")
	}
	t.Cache = make(map[string]*template.Template)
	t.AuthTempl=make(map[string]*template.Template)
	for k, _ := range b.DataList {
		nb := b.GetAll(t.Bucket, k)
		if nb.Error != nil || len(nb.DataList) == 0 {
			continue
		}
		t.loadThisShit(nb.DataList, k)
	}
	return nil
}
func (t *KTemplate) loadThisShit(m map[string][]byte, name string) {
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
	tmpl = template.New(name)
	t.AuthTempl = make(map[string]*template.Template)

	re := regexp.MustCompile("^*[.]html.tpl$")

	for key, value := range m {
		if re.Match([]byte(key)) {
			authTempl[key]=value
		} else {
			var ntmpl *template.Template
			ntmpl = tmpl.New(key)
			_, terr := ntmpl.Parse(string(value))
			if terr != nil {
				panic(terr)
			}
		}
	}
	layout = template.Must(template.New("layout").Funcs(funcs).Parse(string(authTempl["layout.html.tpl"])))
	for key, value := range authTempl {
		if key == "layout.html.tpl" {
			continue
		}
		clone, err := layout.Clone()
		if err != nil {
			panic(err)
		}

		_, err = clone.New("authboss").Funcs(funcMap).Parse(string(value))
		if err != nil {
			panic(err)
		}
		t.AuthTempl[key] = clone
	}

	t.Cache[tmpl.Name()] = tmpl
}
func (t *KTemplate) LoadSingle(name string) error {
	if t.Cache == nil {
		t.Cache = make(map[string]*template.Template)
	}
	if t.Exists(name) {
		return nil
	}
	b := t.Store.GetAll(t.Bucket, name)
	if b.Error != nil {
		return b.Error
	}
	t.loadThisShit(b.DataList, name)
	return nil
}

func (t *KTemplate) Exists(name string) bool {
	tem := t.Cache[name]
	if tem != nil {
		return true
	}
	return false
}
