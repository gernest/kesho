package main

import (
	"net/http"
)

type SessionStorer struct {
	w http.ResponseWriter
	r *http.Request

	SessName string
	Store    *BStore
}

func (s SessionStorer) Get(key string) (string, bool) {
	session, err := s.Store.Get(s.r, s.SessName)
	if err != nil {
		return "", false
	}
	strInf, ok := session.Values[key]
	if !ok {
		return "", false
	}

	str, ok := strInf.(string)
	if !ok {
		return "", false
	}
	return str, true
}

func (s SessionStorer) Put(key, value string) {
	session, err := s.Store.Get(s.r, s.SessName)
	if err != nil {
		return
	}
	session.Values[key] = value
	session.Save(s.r, s.w)
}

func (s SessionStorer) Del(key string) {
	session, err := s.Store.Get(s.r, s.SessName)
	if err != nil {
		return
	}
	delete(session.Values, key)
	session.Save(s.r, s.w)
}
