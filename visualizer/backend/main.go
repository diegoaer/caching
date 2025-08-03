package main

import (
	"encoding/json"
	"net/http"
	"time"

	"caching/lru"
)

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins (you can restrict this if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func cacheHandler(cache *lru.ObservableCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := cache.State()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	}
}

func addToCacheHandler(cache *lru.ObservableCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if payload.Key == "" || payload.Value == "" {
			http.Error(w, "key and value must not be empty", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cache.Cache.Set(payload.Key, payload.Value)
		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	observable := lru.NewObservableCache(5)

	// Add a few example values
	observable.Cache.Set("foo", "bar")
	observable.Cache.SetWithTTL("baz", "qux", time.Minute)

	http.HandleFunc("/cache", withCORS(cacheHandler(observable)))
	http.HandleFunc("/add", withCORS(addToCacheHandler(observable)))
	http.ListenAndServe(":8080", nil)
}
