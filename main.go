package main

import (
	"html/template"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/gorilla/sessions"
)

func main() {
	var (
		mainDB          = "main.db"
		sessDB          = "sessions.db"
		assetsBucket    = "assets"
		templatesBucket = "templates"
		sessionName     = "kesho_"
		sessionBucket   = "sessions"
		secretKey       = "892252c6eade0b4ebf32d94aaed79d20"
		secretValue     = "9451243db34445f4dbf86e0b13bec94d"
	)
	db, _ := bolt.Open(sessDB, 0600, nil)
	defer db.Close()
	defer cleanDB(sessDB)

	// Setup session store
	opts := &sessions.Options{MaxAge: 86400 * 30, Path: "/"}
	ss, err := NewBStoreFromDB(db, sessionBucket, 100, opts, []byte(secretKey), []byte(secretValue))
	if err != nil {
		log.Panic(err)
	}
	mainStorage := NewStorage(mainDB, 0600)
	defer mainStorage.DeleteDatabase()

	// Main app
	app := &Kesho{
		Assets: &Assets{Bucket: assetsBucket, Store: mainStorage},
		Templ: &KTemplate{
			Store:  mainStorage,
			Bucket: templatesBucket,
			Cache:  make(map[string]*template.Template),
		},
		Store:           mainStorage,
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
	if err := app.Templ.LoadToDB("web"); err != nil {
		return err
	}
	return nil
}

func cleanDB(name string) {
	log.Println(os.Remove(name))
}
