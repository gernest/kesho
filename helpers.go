package main

import (
	"html/template"
	"net/http"

	ab "github.com/gernest/authboss"
	"github.com/gorilla/mux"
)

func (k *Kesho) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	k.RenderDefaultView(w, "404.html", nil)
	return
}

func (k *Kesho) ServerProblem(w http.ResponseWriter, msg string) {
	data := make(map[string]interface{})
	data["errorMsg"] = msg
	k.RenderDefaultView(w, "500.html", data)
}

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

	// Posts
	m.HandleFunc("/post/create", k.PostCreate).Methods("POST")
	m.HandleFunc("/post/delete/{slug}/", k.PostDelete)
	m.HandleFunc("/post/update/{slug}", k.PostUpdate)
	m.HandleFunc("/post/view/{slug}/", k.PostView)

	// Views
	m.HandleFunc("/{username}", k.ViewHome)
	m.HandleFunc("/{username}/{slug}", k.ViewPost)
	return m
}

func (k *Kesho) AuthlayoutData(w http.ResponseWriter, r *http.Request) ab.HTMLData {
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

func (k *Kesho) NewSessionStorer(w http.ResponseWriter, r *http.Request) ab.ClientStorer {
	return &SessionStorer{w, r, k.SessionName, k.SessStore}
}

type hdata struct {
	d map[string]interface{}
}

func (h *hdata) Set(key string, v interface{}) {
	h.d[key] = v
}
func (h *hdata) SetSafe(key string, v interface{}) {
	h.d[key] = v
}

func (h *hdata) FlashError(msg string) {
	h.d["flashError"] = template.HTML(msg)
}
func (h *hdata) FlashSuccess(msg string) {
	h.d["flashSuccess"] = template.HTML(msg)
}
func (h *hdata) SetUser(user interface{}) {
	h.d["currentUser"] = user
}
func (h *hdata) StausLogged() {
	h.d["loggedIn"] = true
}
func (h *hdata) Data() map[string]interface{} {
	return h.d
}
func NewHtmlData() *hdata {
	return &hdata{make(map[string]interface{})}
}
