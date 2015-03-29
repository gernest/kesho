package main

import (
	"html/template"
	"log"
	"os"

	"github.com/gorilla/sessions"
	"resenje.org/sessions/boltstore"
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
	opts := &sessions.Options{MaxAge: 400, Path: "/"}
	ss, err := boltstore.NewStoreFromDB(mainStore.db, sessionBucket, 100, opts, []byte(secretKey), []byte(secretValue))
	if err != nil {
		log.Panic(err)
	}

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
		SessionStore:    ss,
		SessionName:     sessionName,
		DefaultTemplate: "kesho",
	}
	app.Templ.Assets = app.Assets
	log.Println(RunMigration(app))
	app.Run()
}

func RunMigration(app *Kesho) error {
	if app.Templ.Exists("leo") {
		return nil
	} else if app.Templ.Exists("kesho") {
		return nil
	} else {
		if err := app.Templ.LoadToDB("leo"); err != nil {
			return err
		}
		if err := app.Templ.LoadToDB("kesho"); err != nil {
			return err
		}
	}
	return nil
}

func cleanDB(name string) {
	log.Println(os.Remove(name))
}
