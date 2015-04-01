package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/justinas/nosurf"
)

func nosurfing(h http.Handler) http.Handler {
	surfing := nosurf.New(h)
	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Failed to validate XSRF Token:", nosurf.Reason(r))
		w.WriteHeader(http.StatusBadRequest)
	}))
	return surfing
}

func (k *Kesho) logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n%s %s %s\n", r.Method, r.URL.Path, r.Proto)
		session, err := k.SessStore.Get(r, sessionCookieName)
		if err == nil {
			fmt.Print("Session: ")
			first := true
			for k, v := range session.Values {
				if first {
					first = false
				} else {
					fmt.Print(", ")
				}
				fmt.Printf("%s = %v", k, v)
			}
			fmt.Println()
		}
		h.ServeHTTP(w, r)
	})
}
