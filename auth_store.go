package main

import (
	"github.com/gernest/authboss"
	"log"
)

type AccountAuth struct {
	Store         *Store
	AccountBucket string
	TokensBucket  string
}

func (a AccountAuth) Create(key string, attr authboss.Attributes) error {
	user := new(Account)
	if err := attr.Bind(user, true); err != nil {
		log.Panicln(err)
		return err
	}
	log.Println("Creating user")
	user.Store = a.Store
	user.Bucket = a.AccountBucket
	return user.StampAndSave()
}

func (a AccountAuth) Put(key string, attr authboss.Attributes) error {
	return a.Create(key, attr)
}

func (a AccountAuth) Get(key string) (result interface{}, err error) {
	acc := NewAccount(a.AccountBucket, a.Store)
	user, err := acc.GetUser(key)
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
	return a.Store.CreateRecord(a.TokensBucket, key, []byte(token)).Error
}

func (a AccountAuth) DelTokens(key string) error {
	return a.Store.RemoveRecord(a.TokensBucket, key).Error
}

func (a AccountAuth) UseToken(givenKey, token string) error {
	a.Store.GetRecord(a.TokensBucket, givenKey)
	if a.Store.Error != nil {
		return authboss.ErrTokenNotFound
	} else if a.Store.Data == nil {
		return authboss.ErrTokenNotFound
	}
	if string(a.Store.Data) == token {
		a.Store.RemoveRecord(a.TokensBucket, givenKey)
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
