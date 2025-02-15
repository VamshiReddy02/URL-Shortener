package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/vamshireddy02/url-shortener/database"
)

// ResolveURL handles redirection from short URL to original URL
func ResolveURL(w http.ResponseWriter, r *http.Request) {
	url := chi.URLParam(r, "url") // Get short URL param

	// Query the database for the original URL
	client := database.CreateClient(0)
	defer client.Close()

	value, err := client.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		http.Error(w, `{"error": "short not found in database"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "cannot connect to DB"}`, http.StatusInternalServerError)
		return
	}

	// Increment the redirect counter
	counterClient := database.CreateClient(1)
	defer counterClient.Close()
	_ = counterClient.Incr(database.Ctx, "counter")

	// Redirect to the original URL
	http.Redirect(w, r, value, http.StatusMovedPermanently)
}
