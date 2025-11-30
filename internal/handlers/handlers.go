package handlers

import (
    "net/http"
    "io"
    "strings"
    "math/rand"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
)

func generateShortID(l int) string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, l) 
    for i := range result {
        result[i] = chars[rand.Intn(len(chars))]
    }
    return string(result)
}

func URLHandler(storage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
			id := strings.TrimPrefix(r.URL.Path, "/")
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
			w.WriteHeader(201)
			w.Write([]byte("http://" + r.Host + "/" + id))
		} else {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	}
}



