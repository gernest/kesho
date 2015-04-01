package main

import (
    "github.com/gernest/authboss"
    "log"
    "net/http"
)

const sessionCookieName = "kesho_"

var sessionStore *BStore

type SessionStorer struct {
    w http.ResponseWriter
    r *http.Request
}

func NewSessionStorer(w http.ResponseWriter, r *http.Request) authboss.ClientStorer {
    return &SessionStorer{w, r}
}

func (s SessionStorer) Get(key string) (string, bool) {
    session, err := sessionStore.Get(s.r, sessionCookieName)
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
    log.Printf("{SUCCESS] session found %s", str)
    return str, true
}

func (s SessionStorer) Put(key, value string) {
    session, err := sessionStore.Get(s.r, sessionCookieName)
    if err != nil {
        return
    }
    session.Values[key] = value
    session.Save(s.r, s.w)
}

func (s SessionStorer) Del(key string) {
    session, err := sessionStore.Get(s.r, sessionCookieName)
    if err != nil {
        return
    }
    delete(session.Values, key)
    session.Save(s.r, s.w)
}
