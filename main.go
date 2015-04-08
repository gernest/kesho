package main

func main() {
	app := NewKesho(nil)
	app.Cleanup()
	app.Run()
}
