package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{
		service: service,
	}
}

func (h *URLHandler) BaseHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {

			id := chi.URLParam(r, "id")
			if id == "" {
				jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			result, err := h.service.BaseGet(r.Context(), id)
			if err != nil {
				switch err {
				case service.ErrURLNotFound:
					jsonError(w, "URL not found", http.StatusNotFound)
				case service.ErrRepository:
					jsonError(w, "Database error", http.StatusInternalServerError)
				default:
					jsonError(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Location", result)
			w.WriteHeader(http.StatusTemporaryRedirect)

		} else if r.Method == http.MethodPost {

			body, err := io.ReadAll(r.Body)
			if err != nil {
				jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			originalURL := string(body)
			if originalURL == "" {
				jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			result, err := h.service.BasePost(r.Context(), originalURL)

			if err != nil {
				switch err {
				case service.ErrInvalidURL:
					jsonError(w, "Invalid URL in request", http.StatusBadRequest)
				case service.ErrURLAlreadyExists:
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(result))
					return
				case service.ErrIDAlreadyExists:
					jsonError(w, "ID already exists", http.StatusConflict)
				case service.ErrRepository:
					jsonError(w, "Database error", http.StatusInternalServerError)
				default:
					jsonError(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(result))

		} else {

			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		}
	})
}

func (h *URLHandler) ShortenHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")

		if !strings.HasPrefix(contentType, "application/json") {
			jsonError(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		var reqs model.RequestShorten

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		result, err := h.service.Shorten(r.Context(), reqs)

		if err != nil {
			switch err {
			case service.ErrInvalidURL:
				jsonError(w, "Invalid URL in request", http.StatusBadRequest)
			case service.ErrURLAlreadyExists:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(result)
				return
			case service.ErrIDAlreadyExists:
				jsonError(w, "ID already exists", http.StatusConflict)
			case service.ErrRepository:
				jsonError(w, "Database error", http.StatusInternalServerError)
			default:
				jsonError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(result); err != nil {
			jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	})
}

func (h *URLHandler) ShortenBatchHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")

		if !strings.HasPrefix(contentType, "application/json") {
			jsonError(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		var reqsBatch []model.RequestShortenBatch

		if err := json.NewDecoder(r.Body).Decode(&reqsBatch); err != nil {
			jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		result, err := h.service.ShortenBatch(r.Context(), reqsBatch)

		if err != nil {
			switch err {
			case service.ErrEmptyRequest:
				jsonError(w, "Request cannot be empty", http.StatusBadRequest)
			case service.ErrInvalidURL:
				jsonError(w, "Invalid URL in request", http.StatusBadRequest)
			case service.ErrMissingCorrelationID:
				jsonError(w, "Missing correlation ID", http.StatusBadRequest)
			case service.ErrFailedToGenerateID:
				jsonError(w, "Failed to generate unique ID", http.StatusInternalServerError)
			case service.ErrIDAlreadyExists:
				jsonError(w, "ID already exists", http.StatusConflict)
			case service.ErrRepository:
				jsonError(w, "Database error", http.StatusInternalServerError)
			default:
				jsonError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if result.CreatedNew {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(result.Responses); err != nil {
			jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	})
}

func (h *URLHandler) PingHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		err := h.service.Ping(r.Context())

		if err != nil {
			jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	})
}

func (h *URLHandler) GetAllHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		result, err := h.service.GetAll(r.Context())
		if err != nil {
			jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(result); err != nil {
			jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	})
}
