package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/vamshireddy02/url-shortener/database"
	"github.com/vamshireddy02/url-shortener/helpers"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

// ShortenURL handles URL shortening requests
func ShortenURL(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var body request
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, `{"error": "cannot parse JSON"}`, http.StatusBadRequest)
		return
	}

	// Implement rate limiting
	client := database.CreateClient(1)
	defer client.Close()

	ip := r.RemoteAddr
	val, err := client.Get(database.Ctx, ip).Result()
	if err == redis.Nil {
		_ = client.Set(database.Ctx, ip, os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := client.TTL(database.Ctx, ip).Result()
			http.Error(w, `{"error": "Rate limit exceeded", "rate_limit_reset": `+strconv.Itoa(int(limit/time.Minute))+`}`, http.StatusServiceUnavailable)
			return
		}
	}

	// Validate the URL
	if !govalidator.IsURL(body.URL) {
		http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
		return
	}

	// Prevent infinite loops by disallowing self-shortening
	if !helpers.RemoveDomainError(body.URL) {
		http.Error(w, `{"error": "haha... nice try"}`, http.StatusServiceUnavailable)
		return
	}

	// Enforce HTTPS
	body.URL = helpers.EnforceHTTP(body.URL)

	// Generate or use custom short URL
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// Store in database
	urlClient := database.CreateClient(0)
	defer urlClient.Close()

	existing, _ := urlClient.Get(database.Ctx, id).Result()
	if existing != "" {
		http.Error(w, `{"error": "URL short already in use"}`, http.StatusForbidden)
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = urlClient.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		http.Error(w, `{"error": "Unable to connect to server"}`, http.StatusInternalServerError)
		return
	}

	// Respond with shortened URL details
	client.Decr(database.Ctx, ip)
	val, _ = client.Get(database.Ctx, ip).Result()
	rateRemaining, _ := strconv.Atoi(val)
	ttl, _ := client.TTL(database.Ctx, ip).Result()

	resp := response{
		URL:             body.URL,
		CustomShort:     os.Getenv("DOMAIN") + "/" + id,
		Expiry:          body.Expiry,
		XRateRemaining:  rateRemaining,
		XRateLimitReset: ttl / time.Minute,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
