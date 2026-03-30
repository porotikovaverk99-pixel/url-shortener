package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/handler"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/middleware"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/service"
	"go.uber.org/zap"
)

const testSecretKey = "test-secret-key"

func setupTestServer() *httptest.Server {
	tmpfile, _ := os.CreateTemp("", "test-*.json")
	tmpfile.Close()

	repo, _ := repository.NewMemoryStorage(tmpfile.Name())

	urlService := service.NewURLService(repo, "http://localhost:8080", 100, 5, 30*time.Second, zap.NewNop())
	h := handler.NewURLHandler(urlService)

	r := chi.NewRouter()

	r.Use(middleware.Auth(testSecretKey))

	r.Method(http.MethodPost, "/", h.BaseHandler())
	r.Method(http.MethodGet, "/{id}", h.BaseHandler())
	r.Method(http.MethodPost, "/api/shorten", h.ShortenHandler())
	r.Method(http.MethodPost, "/api/shorten/batch", h.ShortenBatchHandler())
	r.Method(http.MethodGet, "/ping", h.PingHandler())
	r.Method(http.MethodGet, "/api/user/urls", h.UserUrlsHandler())
	r.Method(http.MethodDelete, "/api/user/urls", h.UserUrlsHandler())

	return httptest.NewServer(r)
}

func ExampleURLHandler_BaseHandler_post() {
	ts := setupTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	// Output:
	// Status: 201
}

func ExampleURLHandler_ShortenHandler() {
	ts := setupTestServer()
	defer ts.Close()

	reqBody := model.RequestShorten{URL: "https://example.com"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	// Output:
	// Status: 201
}

func ExampleURLHandler_ShortenBatchHandler() {
	ts := setupTestServer()
	defer ts.Close()

	reqBody := []model.RequestShortenBatch{
		{CorrelationID: "1", OriginalURL: "https://google.com"},
		{CorrelationID: "2", OriginalURL: "https://yandex.ru"},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var responses []model.ResponseShortenBatch
	json.NewDecoder(resp.Body).Decode(&responses)

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Count:", len(responses))

	// Output:
	// Status: 201
	// Count: 2
}

func ExampleURLHandler_PingHandler() {
	ts := setupTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)

	// Output:
	// Status: 200
}
