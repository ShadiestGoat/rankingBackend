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

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != AUTH_CODE {
			WriteErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}