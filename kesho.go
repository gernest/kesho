package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
	"log"
	"html/template"
	"path/filepath"
	"io/ioutil"
	"net/http"
	"github.com/justinas/nosurf"
	"github.com/astaxie/beego/validation"
	"github.com/gorilla/mux"
	"github.com/monoculum/formam"
	ab "github.com/gernest/authboss"
	_ "github.com/gernest/authboss/register"
	_ "github.com/gernest/authboss/auth"
//	"github.com/justinas/alice"
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

type Kesho struct {
	AccountsBucket  string
	Store           *Store
	Assets          *Assets
	Templ           *KTemplate
	SessStore    *BStore
	SessionName     string
	DefaultTemplate string // The default template for the whole site
}

func (k *Kesho) Auth(w http.ResponseWriter, r *http.Request) {}

func (k *Kesho) Routes() *mux.Router {
	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(k.NotFound)

	m.PathPrefix("/auth").Handler(ab.NewRouter())

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
		sess, err := k.SessStore.New(r, k.SessionName)
		if err != nil {
			log.Println(err)
		}
		sess.Values["username"] = login.UserName
		if err := k.SessStore.Save(r, w, sess); err != nil {
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

func (k *Kesho)Setup() {
	database := AccountAuth{k.Store, k.AccountsBucket, "tokens_"}
	ab.Cfg.Storer = database
	ab.Cfg.OAuth2Storer = database
	ab.Cfg.MountPath = "/auth"
	ab.Cfg.ViewsPath = "ab_views"
	ab.Cfg.LogWriter = os.Stdout
	ab.Cfg.RootURL = `http://localhost:8080`

	ab.Cfg.LayoutDataMaker = layoutData

	b, err := ioutil.ReadFile(filepath.Join("views", "layout.html.tpl"))
	if err != nil {
		panic(err)
	}
	ab.Cfg.Layout = template.Must(template.New("layout").Funcs(funcs).Parse(string(b)))

	ab.Cfg.XSRFName = "csrf_token"
	ab.Cfg.XSRFMaker = func(_ http.ResponseWriter, r *http.Request) string {
		return nosurf.Token(r)
	}
	ab.Cfg.PrimaryID="user_name"

	ab.Cfg.Policies=[]ab.Validator{
		ab.Rules{
			FieldName:"user_name",
			Required:true,
			MinLength:5,
			MaxLength:10,
			AllowWhitespace:false,
		},
		ab.Rules{
			FieldName:"password",
			Required:true,
			MinLength:8,
			MaxLength:20,
			AllowWhitespace:false,
		},
	}
	ab.Cfg.ConfirmFields=[]string{"password", "confirm_password"}

	ab.Cfg.CookieStoreMaker = NewCookieStorer
	ab.Cfg.SessionStoreMaker = NewSessionStorer

	if err := ab.Init(); err != nil {
		log.Fatal(err)
	}
}

func (k Kesho) Run() {
	var (
		httpPort = "8080"
	)
	log.Println("Starting kesho ...")
	log.Println("Loading templates...")
	if err := k.Templ.LoadFromDB(); err != nil {
		log.Fatal(err)
	}

	k.Setup()

	log.Println("done")
	log.Printf("Kesho is running at localhost:%s \n", httpPort)
	addr := fmt.Sprintf(":%s", httpPort)

	//	stack:=alice.New(nosurfing,ab.ExpireMiddleware).Then(k.Routes())

	defer k.Store.Close()
	log.Fatal(http.ListenAndServe(addr, k.Routes()))

}

func layoutData(w http.ResponseWriter, r *http.Request) ab.HTMLData {
	currentUserName := ""
	userInter, err := ab.CurrentUser(w, r)
	if userInter != nil && err == nil {
		currentUserName = userInter.(*Account).UserName
	}

	return ab.HTMLData{
		"loggedin":          userInter != nil,
		"username":          "",
		ab.FlashSuccessKey:  ab.FlashSuccess(w, r),
		ab.FlashErrorKey:    ab.FlashError(w, r),
		"current_user_name": currentUserName,
	}
}