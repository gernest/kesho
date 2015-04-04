package main

import (
	"html/template"
	"log"
	"os"

	"github.com/gorilla/sessions"
)

func main() {
	var (
		mainDB          = "main.db"
		assetsBucket    = "assets"
		templatesBucket = "templates"
		sessionName     = "kesho"
		sessionBucket   = "sessions"
		mainStore       *Store
		secretKey       = "892252c6eade0b4ebf32d94aaed79d20"
		secretValue     = "9451243db34445f4dbf86e0b13bec94d"
	)
	cleanDB(mainDB)

	mainStore = NewStore(mainDB, 0600, nil)

	// Setup session store
	opts := &sessions.Options{MaxAge: 86400 * 30, Path: "/"}
	ss, err := NewBStoreFromDB(mainStore.db, sessionBucket, 100, opts, []byte(secretKey), []byte(secretValue))
	if err != nil {
		log.Panic(err)
	}
	sessionStore = ss

	// Main app
	app := &Kesho{
		Assets: &Assets{Bucket: assetsBucket, Store: mainStore},
		Templ: &KTemplate{
			Store:  mainStore,
			Bucket: templatesBucket,
			Cache:  make(map[string]*template.Template),
		},
		Store:           mainStore,
		AccountsBucket:  "Accounts",
		SessStore:       ss,
		SessionName:     sessionName,
		DefaultTemplate: "kesho",
	}
	app.Templ.Assets = app.Assets
	log.Println(RunMigration(app))
	app.Run()
}

func RunMigration(app *Kesho) error {
	if app.Templ.Exists("kesho") {
		return nil
	} else {
		if err := app.Templ.LoadToDB("kesho"); err != nil {
			return err
		}
	}

	return nil
}

func cleanDB(name string) {
	log.Println(os.Remove(name))
}
