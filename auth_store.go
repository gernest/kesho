package main

import (
	"log"

	"github.com/gernest/authboss"
)

type AccountAuth struct {
	Store         Storage
	AccountBucket string
	TokensBucket  string
}

func (a AccountAuth) Create(key string, attr authboss.Attributes) error {
	user := new(Account)
	if err := attr.Bind(user, true); err != nil {
		log.Panicln(err)
		return err
	}
	user.Store = a.Store
	user.Bucket = a.AccountBucket
	return user.Save()
}

func (a AccountAuth) Put(key string, attr authboss.Attributes) error {
	return a.Create(key, attr)
}

func (a AccountAuth) Get(key string) (result interface{}, err error) {
	user := NewAccount(a.AccountBucket, a.Store)
	user.UserName = key
	err = user.Get()
	if err != nil {
		return nil, authboss.ErrUserNotFound
	}
	return user, nil
}

func (a AccountAuth) PutOAuth(uid, provider string, attr authboss.Attributes) error {
	var user = NewAccount(a.AccountBucket, a.Store)
	if err := attr.Bind(user, true); err != nil {
		return err
	}
	return user.CreateOauth(uid, provider)
}

func (a AccountAuth) GetOAuth(uid, provider string) (result interface{}, err error) {
	var acc = NewAccount(a.AccountBucket, a.Store)
	user, err := acc.GetOauth(uid, provider)
	if err != nil {
		return nil, authboss.ErrUserNotFound
	}
	return user, nil
}

func (a AccountAuth) AddToken(key, token string) error {
	tk := a.Store.CreateDataRecord(a.TokensBucket, key, []byte(token))
	return tk.Error
}

func (a AccountAuth) DelTokens(key string) error {
	tk := a.Store.RemoveDataRecord(a.TokensBucket, key)
	return tk.Error
}

func (a AccountAuth) UseToken(givenKey, token string) error {
	tk := a.Store.GetDataRecord(a.TokensBucket, givenKey)
	if tk.Error != nil {
		return authboss.ErrTokenNotFound
	} else if tk.Data == nil {
		return authboss.ErrTokenNotFound
	}
	if string(a.Store.Data) == token {
		tk.RemoveDataRecord(a.TokensBucket, givenKey)
		return nil
	}
	return authboss.ErrTokenNotFound
}

func (a AccountAuth) ConfirmUser(tok string) (result interface{}, err error) {
	return nil, authboss.ErrUserNotFound
}

func (a AccountAuth) RecoverUser(rec string) (result interface{}, err error) {
	return nil, authboss.ErrUserNotFound
}
