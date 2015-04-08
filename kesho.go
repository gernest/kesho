package main

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	ab "github.com/gernest/authboss"
	_ "github.com/gernest/authboss/auth"
	_ "github.com/gernest/authboss/register"
	_ "github.com/gernest/authboss/remember"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/monoculum/formam"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	VERSION = "0.0.1"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

var funcs = template.FuncMap{
	"formatDate": func(date time.Time) string {
		return date.Format("2006/01/02 03:04pm")
	},
	"yield": func() string { return "" },
}

type Config struct {
	AccountsBucket  string
	TemplatesBucket string
	SessionBucket   string
	SessionName     string
	DefaultTemplate string

	SessionDB string
	MainDB    string

	Secretts []string
}

type Kesho struct {
	AccountsBucket  string
	TemplatesBucket string
	SessionBucket   string

	Store       Storage
	Assets      *Assets
	Templ       *KTemplate
	SessStore   *BStore
	SessionName string

	DefaultTemplate string

	SessionDB string
	MainDB    string
	Secretts  []string

	Cfg *Config
}

func NewKesho(cfg *Config) *Kesho {
	return &Kesho{Cfg: cfg}
}

func (k *Kesho) Initialize() {
	k.initDefaults()
}

func (k *Kesho) initDefaults() {
	k.AccountsBucket = "accounts"
	k.TemplatesBucket = "templates"
	k.SessionBucket = "sessions"
	k.DefaultTemplate = "web"
	k.SessionName = "kesho_"

	kStore := NewStorage("mainstore.db", 0660)
	ass := &Assets{Bucket: "assets", Store: kStore}

	db, err := bolt.Open("sessions.db", 0600, nil)
	if err != nil {
		log.Println(err)
	}
	secrets := []string{
		"892252c6eade0b4ebf32d94aaed79d20",
		"9451243db34445f4dbf86e0b13bec94d",
	}
	opts := &sessions.Options{MaxAge: 86400 * 30, Path: "/"}
	ss, err := NewBStoreFromDB(db, "sessions", 100, opts, []byte(secrets[0]), []byte(secrets[1]))
	if err != nil {
		log.Println(err)
	}
	k.Store = kStore
	k.Assets = ass
	k.Templ = &KTemplate{Store: kStore, Bucket: k.TemplatesBucket, Assets: ass}
	k.SessStore = ss
}

// Our HomePage
func (k *Kesho) HomePage(w http.ResponseWriter, r *http.Request) {
	k.RenderDefaultView(w, "index.html", nil)
	return
}

// Accounts
func (k *Kesho) AccountHome(w http.ResponseWriter, r *http.Request) {
	currentUsr, err := ab.CurrentUser(w, r)
	if err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	if currentUsr == nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}
	usr := currentUsr.(*Account)

	data := NewHtmlData()
	data.Set("Title", "Account")
	data.SetUser(usr)
	data.StausLogged()
	k.RenderDefaultView(w, "accounts/index.html", data.Data())
}

// Posts
func (k *Kesho) PostCreate(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating new post")
	currentUsr, err := ab.CurrentUser(w, r)
	if err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	if currentUsr == nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}
	usr := currentUsr.(*Account)
	data := NewHtmlData()

	r.ParseForm()
	post := new(Post)
	if err = formam.Decode(r.Form, post); err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	post.Account = usr
	err = post.Create()
	if err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	data.SetSafe("Title", usr.UserName)
	data.Set("post", post)
	data.FlashSuccess(post.Title + " imepokelewa na kutangazwa")
	k.RenderDefaultView(w, "accounts/index.html", data.Data())
}

func (k *Kesho) PostUpdate(w http.ResponseWriter, r *http.Request) {
	currentUsr, err := ab.CurrentUser(w, r)
	if err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	if currentUsr == nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	postSlug := vars["slug"]
	if postSlug == "" {
		k.NotFound(w, r)
		return
	}
	usr := currentUsr.(*Account)
	post := new(Post)
	post.Account = usr
	post.Slug = postSlug
	err = post.Get()
	if err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	data := NewHtmlData()
	data.Set("Title", post.Title)
	data.SetUser(usr)
	data.StausLogged()
	data.Set("post", post)
	if r.Method == "POST" {
		r.ParseForm()
		if err = formam.Decode(r.Form, post); err != nil {
			k.ServerProblem(w, err.Error())
			return
		}
		if err = post.Update(); err != nil {
			k.ServerProblem(w, err.Error())
			return
		}
		data.FlashSuccess("Updated " + post.Title)
		k.RenderDefaultView(w, "accounts/index.html", data.Data())
		return
	}
	k.RenderDefaultView(w, "post/update.html", data.Data())
}

func (k *Kesho) PostDelete(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) PostView(w http.ResponseWriter, r *http.Request) {}

// Version
func (k *Kesho) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(VERSION))
	return
}

// Views
func (k *Kesho) ViewHome(w http.ResponseWriter, r *http.Request) {}
func (k *Kesho) ViewPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["username"]
	if uname == "" {
		k.NotFound(w, r)
		return
	}
	pslug := vars["slug"]
	if pslug == "" {
		k.NotFound(w, r)
		return
	}
	user := NewAccount(k.AccountsBucket, k.Store)
	user.UserName = uname
	if err := user.Get(); err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	post := new(Post)
	post.Slug = pslug
	post.Account = user
	if err := post.Get(); err != nil {
		k.ServerProblem(w, err.Error())
		return
	}
	data := NewHtmlData()
	data.Set("user", user)
	data.Set("post", post)
	k.RenderDefaultView(w, "post/post.html", data.Data())
}
func (k *Kesho) ViewSubHome(w http.ResponseWriter, r *http.Request) {}
func (k *Kesho) ViewSubPost(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) RenderDefaultView(w http.ResponseWriter, name string, data interface{}) {
	out := new(bytes.Buffer)
	err := k.Templ.Render(out, k.DefaultTemplate, name, data)

	if err != nil {
		log.Println(err, k.DefaultTemplate, name)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(out.Bytes())
}

func (k *Kesho) Setup() {
	database := AccountAuth{k.Store, k.AccountsBucket, "tokens_"}
	ab.Cfg.Storer = database
	ab.Cfg.OAuth2Storer = database
	ab.Cfg.MountPath = "/auth"
	ab.Cfg.ViewsPath = ""
	ab.Cfg.ResponseTmpl = k.Templ.AuthTempl
	ab.Cfg.LogWriter = os.Stdout
	ab.Cfg.RootURL = `http://localhost:8080`

	ab.Cfg.LayoutDataMaker = k.AuthlayoutData
	ab.Cfg.XSRFName = "csrf_token"
	ab.Cfg.XSRFMaker = func(_ http.ResponseWriter, r *http.Request) string {
		return nosurf.Token(r)
	}
	ab.Cfg.PrimaryID = "user_name"

	ab.Cfg.Policies = []ab.Validator{
		ab.Rules{
			FieldName:       "user_name",
			Required:        true,
			MinLength:       5,
			MaxLength:       10,
			AllowWhitespace: false,
		},
		ab.Rules{
			FieldName:       "password",
			Required:        true,
			MinLength:       8,
			MaxLength:       20,
			AllowWhitespace: false,
		},
	}
	ab.Cfg.ConfirmFields = []string{"password", "confirm_password"}

	ab.Cfg.CookieStoreMaker = k.NewSessionStorer
	ab.Cfg.SessionStoreMaker = k.NewSessionStorer

	if err := ab.Init(); err != nil {
		log.Fatal(err)
	}

	if err := k.Templ.LoadToDB(k.DefaultTemplate); err != nil {
		log.Println(err)
	}
}

func (k Kesho) Run() {
	var (
		httpPort = "8080"
	)
	k.Initialize()
	k.Setup()
	log.Printf("Kesho is running at localhost:%s \n", httpPort)
	addr := fmt.Sprintf(":%s", httpPort)
	log.Fatal(http.ListenAndServe(addr, k.Routes()))

}
func (k *Kesho) Cleanup() error {
	os.Remove(k.MainDB)
	return os.Remove(k.SessionDB)
}
