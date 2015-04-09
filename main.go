package main

import (
	"log"
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
