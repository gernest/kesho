// Package confirm implements confirmation of user registration via e-mail
package confirm

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/gernest/authboss"
	"github.com/gernest/authboss/internal/response"
)

// Storer and FormValue constants
const (
	StoreConfirmToken = "confirm_token"
	StoreConfirmed    = "confirmed"

	FormValueConfirm = "cnf"

	tplConfirmHTML = "confirm_email.html.tpl"
	tplConfirmText = "confirm_email.txt.tpl"
)

var (
	errUserMissing = errors.New("confirm: After registration user must be loaded")
)

// ConfirmStorer must be implemented in order to satisfy the confirm module's
// storage requirements.
type ConfirmStorer interface {
	authboss.Storer
	// ConfirmUser looks up a user by a confirm token. See confirm module for
	// attribute names. If the token is not found in the data store,
	// simply return nil, ErrUserNotFound.
	ConfirmUser(confirmToken string) (interface{}, error)
}

func init() {
	authboss.RegisterModule("confirm", &Confirm{})
}

// Confirm module
type Confirm struct {
	emailHTMLTemplates response.Templates
	emailTextTemplates response.Templates
}

// Initialize the module
func (c *Confirm) Initialize() (err error) {
	var ok bool
	storer, ok := authboss.Cfg.Storer.(ConfirmStorer)
	if storer == nil || !ok {
		return errors.New("confirm: Need a ConfirmStorer")
	}

	c.emailHTMLTemplates, err = response.LoadTemplates(authboss.Cfg.LayoutHTMLEmail, authboss.Cfg.ViewsPath, tplConfirmHTML)
	if err != nil {
		return err
	}
	c.emailTextTemplates, err = response.LoadTemplates(authboss.Cfg.LayoutTextEmail, authboss.Cfg.ViewsPath, tplConfirmText)
	if err != nil {
		return err
	}

	authboss.Cfg.Callbacks.Before(authboss.EventGet, c.beforeGet)
	authboss.Cfg.Callbacks.Before(authboss.EventAuth, c.beforeGet)
	authboss.Cfg.Callbacks.After(authboss.EventRegister, c.afterRegister)

	return nil
}

// Routes for the module
func (c *Confirm) Routes() authboss.RouteTable {
	return authboss.RouteTable{
		"/confirm": c.confirmHandler,
	}
}

// Storage requirements
func (c *Confirm) Storage() authboss.StorageOptions {
	return authboss.StorageOptions{
		authboss.Cfg.PrimaryID: authboss.String,
		authboss.StoreEmail:    authboss.String,
		StoreConfirmToken:      authboss.String,
		StoreConfirmed:         authboss.Bool,
	}
}

func (c *Confirm) beforeGet(ctx *authboss.Context) (authboss.Interrupt, error) {
	if confirmed, err := ctx.User.BoolErr(StoreConfirmed); err != nil {
		return authboss.InterruptNone, err
	} else if !confirmed {
		return authboss.InterruptAccountNotConfirmed, nil
	}

	return authboss.InterruptNone, nil
}

// AfterRegister ensures the account is not activated.
func (c *Confirm) afterRegister(ctx *authboss.Context) error {
	if ctx.User == nil {
		return errUserMissing
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	sum := md5.Sum(token)

	ctx.User[StoreConfirmToken] = base64.StdEncoding.EncodeToString(sum[:])

	if err := ctx.SaveUser(); err != nil {
		return err
	}

	email, err := ctx.User.StringErr(authboss.StoreEmail)
	if err != nil {
		return err
	}

	goConfirmEmail(c, email, base64.URLEncoding.EncodeToString(token))

	return nil
}

var goConfirmEmail = func(c *Confirm, to, token string) {
	go c.confirmEmail(to, token)
}

// confirmEmail sends a confirmation e-mail.
func (c *Confirm) confirmEmail(to, token string) {
	p := path.Join(authboss.Cfg.MountPath, "confirm")
	url := fmt.Sprintf("%s%s?%s=%s", authboss.Cfg.RootURL, p, url.QueryEscape(FormValueConfirm), url.QueryEscape(token))

	email := authboss.Email{
		To:      []string{to},
		From:    authboss.Cfg.EmailFrom,
		Subject: authboss.Cfg.EmailSubjectPrefix + "Confirm New Account",
	}

	err := response.Email(email, c.emailHTMLTemplates, tplConfirmHTML, c.emailTextTemplates, tplConfirmText, url)
	if err != nil {
		fmt.Fprintf(authboss.Cfg.LogWriter, "confirm: Failed to send e-mail: %v", err)
	}
}

func (c *Confirm) confirmHandler(ctx *authboss.Context, w http.ResponseWriter, r *http.Request) error {
	token, err := ctx.FirstFormValueErr(FormValueConfirm)
	if err != nil {
		return err
	}

	toHash, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return authboss.ErrAndRedirect{
			Location: "/", Err: fmt.Errorf("confirm: token failed to decode %q => %v\n", token, err),
		}
	}

	sum := md5.Sum(toHash)

	dbTok := base64.StdEncoding.EncodeToString(sum[:])
	user, err := authboss.Cfg.Storer.(ConfirmStorer).ConfirmUser(dbTok)
	if err == authboss.ErrUserNotFound {
		return authboss.ErrAndRedirect{Location: "/", Err: errors.New("confirm: token not found")}
	} else if err != nil {
		return err
	}

	ctx.User = authboss.Unbind(user)

	ctx.User[StoreConfirmToken] = ""
	ctx.User[StoreConfirmed] = true

	key, err := ctx.User.StringErr(authboss.Cfg.PrimaryID)
	if err != nil {
		return err
	}

	if err := ctx.SaveUser(); err != nil {
		return err
	}

	ctx.SessionStorer.Put(authboss.SessionKey, key)
	response.Redirect(ctx, w, r, authboss.Cfg.RegisterOKPath, "You have successfully confirmed your account.", "", true)

	return nil
}
