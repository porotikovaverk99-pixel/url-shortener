package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	packgzip "github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
	storage "github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLHandler(t *testing.T) {
	type want struct {
		statusCode  int
		response    string
		contentType string
		location    string
	}

	tests := []struct {
		name    string
		url     string
		method  string
		body    io.Reader
		storage storage.URLStorage
		baseURL string
		want    want
	}{
		{
			name:    "test POST",
			url:     "/",
			method:  http.MethodPost,
			body:    strings.NewReader("google.ru"),
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  201,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:    "test POST url exists",
			url:     "/",
			method:  http.MethodPost,
			body:    strings.NewReader("google.ru"),
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  409,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:    "test POST no data",
			url:     "/",
			method:  http.MethodPost,
			body:    strings.NewReader(""),
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:    "test GET",
			url:     "/abcdef",
			method:  http.MethodGet,
			body:    nil,
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  307,
				contentType: "",
				location:    "yandex.ru",
			},
		},
		{
			name:    "test GET no url",
			url:     "/aaaaaa",
			method:  http.MethodGet,
			body:    nil,
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  404,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:    "test GET no id",
			url:     "/",
			method:  http.MethodGet,
			body:    nil,
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:    "test PUT",
			url:     "/",
			method:  http.MethodPut,
			body:    nil,
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  405,
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*.json")
			require.NoError(t, err)
			tmpfile.Close()
			defer os.Remove(tmpfile.Name())

			storage, err := storage.NewMemoryStorage(tmpfile.Name())
			require.NoError(t, err)

			if test.name == "test POST url exists" {
				storage.Save(context.Background(), "abcdef", "google.ru")
			}
			if test.name == "test GET" {
				storage.Save(context.Background(), "abcdef", "yandex.ru")
			}

			r := chi.NewRouter()
			handler := URLHandler(storage, test.baseURL)
			r.Post("/", handler.ServeHTTP)
			r.Get("/{id}", handler.ServeHTTP)
			r.HandleFunc("/*", handler.ServeHTTP)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(test.method, ts.URL+test.url, test.body)
			require.NoError(t, err)

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			res, err := client.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)
			if test.method == http.MethodPost && (res.StatusCode == 201 || res.StatusCode == 200) {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				responseURL := string(body)
				parts := strings.Split(responseURL, "/")
				short := parts[len(parts)-1]
				original, err := storage.Get(context.Background(), short)
				assert.NoError(t, err)
				assert.Equal(t, "google.ru", original)
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}

func TestURLHandlerShorten(t *testing.T) {
	type want struct {
		statusCode  int
		response    string
		contentType string
		location    string
	}

	tests := []struct {
		name    string
		url     string
		method  string
		body    io.Reader
		headers map[string]string
		baseURL string
		want    want
	}{
		{
			name:   "test POST",
			url:    "/api/shorten",
			method: http.MethodPost,
			body:   strings.NewReader(`{"url": "google.ru"}`),
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  201,
				contentType: "application/json",
				location:    "",
			},
		},
		{
			name:   "test POST url exists",
			url:    "/api/shorten",
			method: http.MethodPost,
			body:   strings.NewReader(`{"url": "google.ru"}`),
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  409,
				contentType: "application/json",
				location:    "",
			},
		},
		{
			name:    "test Get",
			url:     "/api/shorten",
			method:  http.MethodGet,
			body:    nil,
			headers: map[string]string{},
			baseURL: "http://localhost:8080",
			want: want{
				statusCode:  405,
				contentType: "",
				location:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*.json")
			require.NoError(t, err)
			tmpfile.Close()
			defer os.Remove(tmpfile.Name())

			storage, err := storage.NewMemoryStorage(tmpfile.Name())
			require.NoError(t, err)

			if test.name == "test POST url exists" {
				storage.Save(context.Background(), "abcdef", "google.ru")
			}

			r := chi.NewRouter()
			handler := URLHandlerShorten(storage, test.baseURL)
			r.Post("/api/shorten", handler.ServeHTTP)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(test.method, ts.URL+test.url, test.body)

			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			require.NoError(t, err)

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			res, err := client.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if res.StatusCode == 200 || res.StatusCode == 201 {

				var ress ResponseShorten
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				err = json.Unmarshal(body, &ress)
				require.NoError(t, err)

				responseURL := ress.Result

				parts := strings.Split(responseURL, "/")
				short := parts[len(parts)-1]
				original, err := storage.Get(context.Background(), short)
				assert.NoError(t, err)
				assert.Equal(t, "google.ru", original)
			}

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}

func TestGzipCompression(t *testing.T) {

	tmpfile, err := os.CreateTemp("", "test-*.json")
	require.NoError(t, err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	storage, err := storage.NewMemoryStorage(tmpfile.Name())
	require.NoError(t, err)

	storage.Save(context.Background(), "abcdef", "google.ru")

	r := chi.NewRouter()
	handler := URLHandlerShorten(storage, "http://localhost:8080")

	handlerMiddleware := packgzip.GzipMiddleware(handler)

	r.Post("/api/shorten", handlerMiddleware.ServeHTTP)

	srv := httptest.NewServer(r)
	defer srv.Close()

	requestBody := `{
        "url": "google.ru"
    }`

	successBody := `{
        "result": "http://localhost:8080/abcdef"
    }`

	t.Run("sends_gzip", func(t *testing.T) {

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/shorten", buf)

		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "")

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Do(req)
		require.NoError(t, err)

		defer res.Body.Close()

		require.Equal(t, http.StatusConflict, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))

	})

	t.Run("accepts_gzip", func(t *testing.T) {

		buf := bytes.NewBufferString(requestBody)
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/shorten", buf)

		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Do(req)

		require.NoError(t, err)
		require.Equal(t, http.StatusConflict, res.StatusCode)

		defer res.Body.Close()

		zr, err := gzip.NewReader(res.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))

	})
}

func TestURLHandlerShortenBatch(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "test POST batch success",
			method: http.MethodPost,
			body: `[
				{"correlation_id": "1", "original_url": "https://google.com"},
				{"correlation_id": "2", "original_url": "https://yandex.ru"}
			]`,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:   "test GET method not allowed",
			method: http.MethodGet,
			body:   "",
			want: want{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*.json")
			require.NoError(t, err)
			tmpfile.Close()
			defer os.Remove(tmpfile.Name())

			storage, err := storage.NewMemoryStorage(tmpfile.Name())
			require.NoError(t, err)

			r := chi.NewRouter()
			handler := URLHandlerShortenBatch(storage, "http://localhost:8080")
			r.Post("/api/shorten/batch", handler.ServeHTTP)
			r.HandleFunc("/api/shorten/batch", handler.ServeHTTP)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(test.method, ts.URL+"/api/shorten/batch", strings.NewReader(test.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestURLHandlerPing(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   int
	}{
		{
			name:   "test GET ping success",
			method: http.MethodGet,
			want:   http.StatusOK,
		},
		{
			name:   "test POST method not allowed",
			method: http.MethodPost,
			want:   http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*.json")
			require.NoError(t, err)
			tmpfile.Close()
			defer os.Remove(tmpfile.Name())

			storage, err := storage.NewMemoryStorage(tmpfile.Name())
			require.NoError(t, err)

			r := chi.NewRouter()
			handler := URLHandlerPing(storage)
			r.Get("/ping", handler.ServeHTTP)
			r.HandleFunc("/ping", handler.ServeHTTP)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(test.method, ts.URL+"/ping", nil)
			require.NoError(t, err)
			client := &http.Client{}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.want, res.StatusCode)
		})
	}
}
