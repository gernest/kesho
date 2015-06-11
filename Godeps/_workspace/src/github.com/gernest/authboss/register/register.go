// Package register allows for user registration.
package register

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"github.com/gernest/authboss"
	"github.com/gernest/authboss/internal/response"
)

const (
	tplRegister = "register.html.tpl"
)

// RegisterStorer must be implemented in order to satisfy the register module's
// storage requirments.
type RegisterStorer interface {
	authboss.Storer
	// Create is the same as put, except it refers to a non-existent key.
	Create(key string, attr authboss.Attributes) error
}

func init() {
	authboss.RegisterModule("register", &Register{})
}

// Register module.
type Register struct {
	templates response.Templates
}

// Initialize the module.
func (r *Register) Initialize() (err error) {
	if authboss.Cfg.Storer == nil {
		return errors.New("register: Need a RegisterStorer")
	}

	if _, ok := authboss.Cfg.Storer.(RegisterStorer); !ok {
		return errors.New("register: RegisterStorer required for register functionality")
	}

	if r.templates, err = response.LoadTemplates(authboss.Cfg.Layout, authboss.Cfg.ViewsPath, tplRegister); err != nil {
		return err
	}

	return nil
}

// Routes creates the routing table.
func (r *Register) Routes() authboss.RouteTable {
	return authboss.RouteTable{
		"/register": r.registerHandler,
	}
}

// Storage returns storage requirements.
func (r *Register) Storage() authboss.StorageOptions {
	return authboss.StorageOptions{
		authboss.Cfg.PrimaryID: authboss.String,
		authboss.StorePassword: authboss.String,
	}
}

func (reg *Register) registerHandler(ctx *authboss.Context, w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		data := authboss.HTMLData{
			"primaryID":      authboss.Cfg.PrimaryID,
			"primaryIDValue": "",
		}
		return reg.templates.Render(ctx, w, r, tplRegister, data)
	case "POST":
		return reg.registerPostHandler(ctx, w, r)
	}
	return nil
}

func (reg *Register) registerPostHandler(ctx *authboss.Context, w http.ResponseWriter, r *http.Request) error {
	key, _ := ctx.FirstPostFormValue(authboss.Cfg.PrimaryID)
	password, _ := ctx.FirstPostFormValue(authboss.StorePassword)

	policies := authboss.FilterValidators(authboss.Cfg.Policies, authboss.Cfg.PrimaryID, authboss.StorePassword)
	validationErrs := ctx.Validate(policies, authboss.Cfg.ConfirmFields...)

	if len(validationErrs) != 0 {
		data := authboss.HTMLData{
			"primaryID":      authboss.Cfg.PrimaryID,
			"primaryIDValue": key,
			"errs":           validationErrs.Map(),
		}

		return reg.templates.Render(ctx, w, r, tplRegister, data)
	}

	attr, err := ctx.Attributes() // Attributes from overriden forms
	if err != nil {
		return err
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(password), authboss.Cfg.BCryptCost)
	if err != nil {
		return err
	}

	attr[authboss.Cfg.PrimaryID] = key
	attr[authboss.StorePassword] = string(pass)
	ctx.User = attr

	if err := authboss.Cfg.Storer.(RegisterStorer).Create(key, attr); err != nil {
		return err
	}

	if err := authboss.Cfg.Callbacks.FireAfter(authboss.EventRegister, ctx); err != nil {
		return err
	}

	if authboss.IsLoaded("confirm") {
		response.Redirect(ctx, w, r, authboss.Cfg.RegisterOKPath, "Account successfully created, please verify your e-mail address.", "", true)
		return nil
	}

	ctx.SessionStorer.Put(authboss.SessionKey, key)
	response.Redirect(ctx, w, r, authboss.Cfg.RegisterOKPath, "Account successfully created, you are now logged in.", "", true)

	return nil
}
