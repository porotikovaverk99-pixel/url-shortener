package handlers

import (
    "net/http"
    "io"
    "math/rand"
	strg "github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"errors"
)

func generateShortID(l int) string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, l) 
    for i := range result {
        result[i] = chars[rand.Intn(len(chars))]
    }
    return string(result)
}

func URLHandler(storage strg.URLStorage, baseURL string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {

			id := chi.URLParam(r, "id")
			if id == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			originalURL, err := storage.Get(r.Context(), id)
			if err != nil {
				if errors.Is(err, strg.ErrURLNotFound) {
					http.Error(w, "Not found", http.StatusNotFound)
				} else {
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)

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
			err = storage.Save(r.Context(), id, originalURL)
			if err != nil {
				if errors.Is(err, strg.ErrURLAlreadyExists) {
					http.Error(w, "Conflict", http.StatusConflict)
				} else {
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(baseURL + "/" + id))

		} else {

			http.Error(w, "Bad Request", http.StatusBadRequest)

		} 
	} 
}

func URLHandlerShorten(storage strg.URLStorage, baseURL string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {

			id := chi.URLParam(r, "id")
			if id == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			originalURL, err := storage.Get(r.Context(), id)
			if err != nil {
				if errors.Is(err, strg.ErrURLNotFound) {
					http.Error(w, "Not found", http.StatusNotFound)
				} else {
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)

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
			err = storage.Save(r.Context(), id, originalURL)
			if err != nil {
				if errors.Is(err, strg.ErrURLAlreadyExists) {
					http.Error(w, "Conflict", http.StatusConflict)
				} else {
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(baseURL + "/" + id))

		} else {

			http.Error(w, "Bad Request", http.StatusBadRequest)

		} 
	} 
}




