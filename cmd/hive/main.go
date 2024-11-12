package main

import "github.com/tendant/chi-demo/app"

func main() {
	server := app.DefaultApp()
	server.Run()
}
