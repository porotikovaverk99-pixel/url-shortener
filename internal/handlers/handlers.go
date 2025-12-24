package handlers

import (
    "net/http"
    "io"
    "math/rand"
	strg "github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"encoding/json"
	"errors"
	"context"
)

type RequestShorten struct {
	Url string `json:"url"`
}

type ResponseShorten struct {
	Result string `json:"result"`
}

func generateShortID(l int) string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, l) 
    for i := range result {
        result[i] = chars[rand.Intn(len(chars))]
    }
    return string(result)
}

func processURL(ctx context.Context, storage strg.URLStorage, url string) (string, string, int) {
	
	var id string
	var status int

	foundID, err := storage.FindIDByURL(ctx, url)

	if err == nil {

		status = http.StatusOK
		id = foundID

	} else {

		if !errors.Is(err, strg.ErrIDNotFound) {
			return "", "Internal Server Error", http.StatusInternalServerError
		}

		id = generateShortID(8)
		err := storage.Save(ctx, id, url)

		if err != nil {
			if errors.Is(err, strg.ErrIDAlreadyExists) {
				return "", "Conflict", http.StatusConflict
			} else {
				return "", "Internal Server Error", http.StatusInternalServerError
			}
		}
		status = http.StatusCreated
	}

	return id, "", status
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
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
			id, statusText, status := processURL(r.Context(), storage, originalURL)

			if status != http.StatusOK && status != http.StatusCreated {
				http.Error(w, statusText, status)
				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(status)
			w.Write([]byte(baseURL + "/" + id))

		} else {

			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)

		} 
	} 
}

func URLHandlerShorten(storage strg.URLStorage, baseURL string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")

		if contentType != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}

        var reqs RequestShorten

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if reqs.Url == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		id, statusText, status := processURL(r.Context(), storage, reqs.Url)

		if status != http.StatusOK && status != http.StatusCreated {
			http.Error(w, statusText, status)
			return
		}

		ress := ResponseShorten{
			Result: baseURL + "/" + id,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if err := json.NewEncoder(w).Encode(ress); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	} 
}





