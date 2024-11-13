package main

import (
	"github.com/tendant/chi-demo/app"
	"github.com/tendant/opa-board/board"
)

func main() {
	server := app.DefaultApp()

	server.R.Get("/", board.RenderIndex)
	server.R.Post("/evaluate", board.EvaluatePolicy)

	server.Run()
}
