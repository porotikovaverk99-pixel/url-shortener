package handlers

import (
    "net/http"
    "io"
    "math/rand"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

func generateShortID(l int) string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, l) 
    for i := range result {
        result[i] = chars[rand.Intn(len(chars))]
    }
    return string(result)
}

func URLHandler(storage storage.URLStorage, baseURL string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {

			id := chi.URLParam(r, "id")
			if id == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			originalURL, err := storage.Get(id)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Location", originalURL)
			w.WriteHeader(307)

		} else if r.Method == http.MethodPost {

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			originalURL := string(body)
			if originalURL == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			id := generateShortID(8)
			err = storage.Save(id, originalURL)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			 w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(201)
			w.Write([]byte(baseURL + "/" + id))

		} else {

			http.Error(w, "Bad Request", http.StatusBadRequest)

		}
	} 
}



