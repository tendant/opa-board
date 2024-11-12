package main

import (
	"github.com/tendant/chi-demo/app"
	"github.com/tendant/opa-hive/hive"
)

func main() {
	server := app.DefaultApp()

	server.R.Get("/", hive.RenderIndex)
	server.R.Post("/evaluate", hive.EvaluatePolicy)

	server.Run()
}
