package main

import (
	"net/http"
	"io"
	"math/rand"
	"strings"
)

var urlStorage = map[string]string{}
var lenID int = 6

func urlAction(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		originalURL := urlStorage[id]
		if originalURL == "" {
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
		id := generateShortID(lenID)
		urlStorage[id] = originalURL
		w.WriteHeader(201)
		w.Write([]byte("http://" + r.Host + "/" + id))
	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

func generateShortID(l int) string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, l) 
    for i := range result {
        result[i] = chars[rand.Intn(len(chars))]
    }
    return string(result)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", urlAction)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
