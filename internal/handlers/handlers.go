package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	strg "github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
)

type RequestShorten struct {
	URL string `json:"url"`
}

type RequestShortenBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseShorten struct {
	Result string `json:"result"`
}

type ResponseShortenBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
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

	foundID, err := storage.FindIDByURL(ctx, url)

	if err == nil {

		return foundID, "", http.StatusOK

	} else {

		if !errors.Is(err, strg.ErrIDNotFound) {
			return "", http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
		}

		id := generateShortID(8)
		err := storage.Save(ctx, id, url)

		if err != nil {
			if errors.Is(err, strg.ErrIDAlreadyExists) {
				return "", http.StatusText(http.StatusConflict), http.StatusConflict
			} else {
				return "", http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
			}
		}

		return id, "", http.StatusCreated

	}

}

func URLHandler(storage strg.URLStorage, baseURL string) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {

			id := chi.URLParam(r, "id")
			if id == "" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			originalURL, err := storage.Get(r.Context(), id)
			if err != nil {
				if errors.Is(err, strg.ErrURLNotFound) {
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				} else {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)

		} else if r.Method == http.MethodPost {

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			originalURL := string(body)
			if originalURL == "" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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

			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		}
	})
}

func URLHandlerShorten(storage strg.URLStorage, baseURL string) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")

		if !strings.HasPrefix(contentType, "application/json") {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		var reqs RequestShorten

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if reqs.URL == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		id, statusText, status := processURL(r.Context(), storage, reqs.URL)

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
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	})
}

func URLHandlerShortenBatch(storage strg.URLStorage, baseURL string) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")

		if !strings.HasPrefix(contentType, "application/json") {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		var reqsBatch []RequestShortenBatch

		if err := json.NewDecoder(r.Body).Decode(&reqsBatch); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if len(reqsBatch) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		urls := make([]string, 0, len(reqsBatch))

		for _, item := range reqsBatch {
			if item.OriginalURL == "" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			urls = append(urls, item.OriginalURL)
		}

		foundIDs, err := storage.FindIDByURLs(r.Context(), urls)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		batch := []strg.BatchItem{}
		ressBatch := make([]ResponseShortenBatch, 0, len(reqsBatch))
		generatedIDs := make(map[string]bool)

		for _, item := range reqsBatch {
			if foundID, ok := foundIDs[item.OriginalURL]; ok {
				ressBatch = append(ressBatch, ResponseShortenBatch{CorrelationID: item.CorrelationID, ShortURL: baseURL + "/" + foundID})
			} else {
				generatedID := generateShortID(8)
				attempts := 0
				for generatedIDs[generatedID] && attempts < 100 {
					generatedID = generateShortID(8)
					attempts++
				}
				if attempts >= 100 {
					http.Error(w, "Failed to generate unique ID", http.StatusInternalServerError)
					return
				}
				batch = append(batch, strg.BatchItem{ShortURL: generatedID, OriginalURL: item.OriginalURL})
				ressBatch = append(ressBatch, ResponseShortenBatch{CorrelationID: item.CorrelationID, ShortURL: baseURL + "/" + generatedID})
			}
		}

		err = storage.SaveBatch(r.Context(), batch)
		if err != nil {
			if errors.Is(err, strg.ErrIDAlreadyExists) {
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		var status int = http.StatusCreated
		if len(foundIDs) == len(reqsBatch) {
			status = http.StatusOK
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if err := json.NewEncoder(w).Encode(ressBatch); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	})
}

func URLHandlerPing(storage strg.URLStorage) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		err := storage.Ping(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	})
}
