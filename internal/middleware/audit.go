package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/porotikovaverk99-pixel/url-shortener/internal/audit"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func AuditMiddleware(manager *audit.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			next.ServeHTTP(rw, r)

			if rw.statusCode >= 200 && rw.statusCode < 400 {
				action := determineAction(r.Method, r.URL.Path)

				if action == "" {
					return
				}
				var originalURL string

				if action == "shorten" {
					originalURL = extractOriginalURLFromBody(r.URL.Path, bodyBytes)
				} else if action == "follow" {
					originalURL = rw.Header().Get("Location")
				}
				if originalURL != "" {
					event := audit.Event{
						Ts:     time.Now().Unix(),
						Action: action,
						UserID: getUserID(r),
						URL:    originalURL,
					}
					manager.Notify(event)
				}
			}
		})
	}
}

func extractOriginalURLFromBody(path string, bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return ""
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err == nil && req.URL != "" {
		return req.URL
	}

	return strings.TrimSpace(string(bodyBytes))
}

func determineAction(method, path string) string {

	if method == http.MethodPost {
		return "shorten"
	} else {
		return "follow"
	}

}

func getUserID(r *http.Request) string {
	if userID, ok := r.Context().Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
