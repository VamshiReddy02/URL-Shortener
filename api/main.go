package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/vamshireddy02/url-shortener/routes"
)

// setupRoutes defines the URL shortening and resolving routes
func setupRoutes(r *chi.Mux) {
	r.Post("/api/v1", routes.ShortenURL)
	r.Get("/{url}", routes.ResolveURL)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	setupRoutes(r)

	port := 3000

	fmt.Println("Server is running on port", port)
	log.Fatal(http.ListenAndServe(":3000", r))
}
