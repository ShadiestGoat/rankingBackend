package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func Router() http.Handler {
	r := chi.NewRouter()
	
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
	}))

	r.Mount("/match", matchRouter())

	r.Mount("/players", playerRouter())

	r.HandleFunc("/ws", socketHandler)

	return r
}
