package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/astaxie/beego/validation"
	"github.com/e-dard/netbug"
	"github.com/gorilla/mux"
	"github.com/monoculum/formam"
	"resenje.org/sessions/boltstore"
)

const (
	VERSION = "0.0.1"
)

type Kesho struct {
	AccountsBucket  string
	Store           *Store
	Assets          *Assets
	Templ           *KTemplate
	SessionStore    *boltstore.Store
	SessionName     string
	DefaultTemplate string // The default template for the whole site
}

func (k *Kesho) Auth(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) Routes() *mux.Router {
	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(k.NotFound)

	// Home page
	m.HandleFunc("/", k.HomePage)

	// Static Assets
	m.HandleFunc("/static/{filename:.*}", k.Assets.Serve)

	// Accounts
	m.HandleFunc("/accounts", k.AccountHome)
	m.HandleFunc("/accounts/create", k.AccountCreate)
	m.HandleFunc("/accounts/login", k.AccountLogin)
	m.HandleFunc("/accounts/delete/{username}", k.AccountDelete)
	m.HandleFunc("/accounts/update/{username}", k.AccountUpdate)

	// Posts
	m.HandleFunc("/post/create/", k.PostCreate)
	m.HandleFunc("/post/delete/{slug}/", k.PostDelete)
	m.HandleFunc("/post/update/{slug}", k.PostUpdate)
	m.HandleFunc("/post/view/{slug}/", k.PostView)

	// Version
	m.HandleFunc("/version", k.Version)
	// Views
	m.HandleFunc("/{username}", k.ViewHome)
	m.HandleFunc("/{username}/{slug}", k.ViewPost)

	// Subdomain View
	s := m.Host("{subdomain:[a-z]+}.domain.com").Subrouter()
	s.HandleFunc("/", k.ViewSubHome)
	s.HandleFunc("/{slug}", k.ViewSubPost)
	return m
}

// Our HomePage
func (k *Kesho) HomePage(w http.ResponseWriter, r *http.Request) {
	k.RenderDefaultView(w, "index.html", nil)
	return
}

// Accounts
func (k *Kesho) AccountHome(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) AccountCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		k.RenderDefaultView(w, "accounts/new.html", nil)
		return
	} else if r.Method == "POST" {
		user := NewAccount(k.AccountsBucket, k.Store)
		valid := &validation.Validation{}
		data := make(map[string]interface{})
		m := make(map[string]string)
		r.ParseForm()

		if err := formam.Decode(r.Form, user); err != nil {
			m["Some Fish"] = err.Error()
			data["errors"] = m
			k.RenderDefaultView(w, "accounts/new.html", data)
			return
		}
		b, err := valid.Valid(user)
		if err != nil || !b {
			for k, v := range valid.ErrorsMap {
				m[k] = v.Message
			}
			data["errors"] = m
			k.RenderDefaultView(w, "accounts/new.html", data)
			return
		}
		if user.IsUser() {
			m["Some Fish"] = "The name of the blog has already been taken"
			data["errors"] = m
			k.RenderDefaultView(w, "accounts/new.html", data)
			return
		}
		if err := user.CreateUser(); err != nil {
			m["Some Fish"] = err.Error()
			data["errors"] = m
			k.RenderDefaultView(w, "accounts/new.html", data)
			return
		}
		http.Redirect(w, r, "/accounts/login", http.StatusFound)
		return
	}
}

func (k *Kesho) AccountUpdate(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) AccountDelete(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) AccountLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		k.RenderDefaultView(w, "accounts/login.html", nil)
		return
	} else if r.Method == "POST" {
		login := NewAccount(k.AccountsBucket, k.Store)
		valid := &validation.Validation{}
		r.ParseForm()
		if err := formam.Decode(r.Form, login); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b, err := valid.Valid(login)
		if err != nil || !b {
			log.Println(valid.ErrorsMap)
			data := make(map[string]interface{})
			m := make(map[string]string)
			for k, v := range valid.ErrorsMap {
				m[k] = v.Message
			}
			data["errors"] = m
			k.RenderDefaultView(w, "accounts/login.html", data)
			return
		}
		if err = login.login(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sess, err := k.SessionStore.New(r, k.SessionName)
		if err != nil {
			log.Println(err)
		}
		sess.Values["username"] = login.UserName
		if err := k.SessionStore.Save(r, w, sess); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Posts
func (k *Kesho) PostCreate(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) PostUpdate(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) PostDelete(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) PostView(w http.ResponseWriter, r *http.Request) {}

// Version
func (k *Kesho) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(VERSION))
	return
}

// Views
func (k *Kesho) ViewHome(w http.ResponseWriter, r *http.Request)    {}
func (k *Kesho) ViewPost(w http.ResponseWriter, r *http.Request)    {}
func (k *Kesho) ViewSubHome(w http.ResponseWriter, r *http.Request) {}
func (k *Kesho) ViewSubPost(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) RenderDefaultView(w http.ResponseWriter, name string, data interface{}) {
	out := new(bytes.Buffer)
	err := k.Templ.Render(out, k.DefaultTemplate, name, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(out.Bytes())
}

func (k *Kesho) NotFound(w http.ResponseWriter, r *http.Request) {
	k.RenderDefaultView(w, "404.html", nil)
	return
}

func (k *Kesho) InternalProblem(w http.ResponseWriter) {}

func (k Kesho) Run() {
	var (
		httpPort = "8080"
	)
	log.Println("Starting kesho ...")
	log.Println("Loading templates...")
	if err := k.Templ.LoadFromDB(); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
	log.Printf("Kesho is running at localhost:%s \n", httpPort)
	addr := fmt.Sprintf(":%s", httpPort)

	defer k.Store.Close()

	smux := http.NewServeMux()
	netbug.RegisterHandler("/profile", smux)
	smux.Handle("/", k.Routes())
	log.Fatal(http.ListenAndServe(addr, smux))

}
