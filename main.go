package main

import (
	"log"
	"os"
)

func main() {
	app := NewKesho(nil)
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
