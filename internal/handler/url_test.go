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
	"time"

	"github.com/go-chi/chi/v5"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/auth"
	packgzip "github.com/porotikovaverk99-pixel/url-shortener/internal/gzip"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/repository"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testSecretKey = "test-secret-key-for-tests"

type contextKey string

const userIDContextKey contextKey = "userID"

func setupTest() (*URLHandler, repository.URLRepository, func(), error) {
	tmpfile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		return nil, nil, nil, err
	}
	tmpfile.Close()

	repo, err := repository.NewMemoryStorage(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		return nil, nil, nil, err
	}

	urlService := service.NewURLService(repo, "http://localhost:8080", 100, 5, 30*time.Second, zap.NewNop())
	handler := NewURLHandler(urlService)

	cleanup := func() {
		os.Remove(tmpfile.Name())
	}

	return handler, repo, cleanup, nil
}

func generateTestToken(userID, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
	})
	return token.SignedString([]byte(secretKey))
}

func newRequestWithUserID(method, url, body string) (*http.Request, error) {
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	token, err := generateTestToken("test-user-id", testSecretKey)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}

func TestURLHandler_BaseHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		location    string
	}

	tests := []struct {
		name   string
		url    string
		method string
		body   string
		setup  func(repo repository.URLRepository)
		want   want
	}{
		{
			name:   "test POST success",
			url:    "/",
			method: http.MethodPost,
			body:   "google.ru",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "test POST url exists",
			url:    "/",
			method: http.MethodPost,
			body:   "google.ru",
			setup: func(repo repository.URLRepository) {
				ctx := context.WithValue(context.Background(), userIDContextKey, "test-user-id")
				repo.Save(ctx, "abcdef", "google.ru", "test-user-id")
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "test POST no data",
			url:    "/",
			method: http.MethodPost,
			body:   "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:   "test GET success",
			url:    "/abcdef",
			method: http.MethodGet,
			setup: func(repo repository.URLRepository) {
				ctx := context.WithValue(context.Background(), userIDContextKey, "test-user-id")
				repo.Save(ctx, "abcdef", "yandex.ru", "test-user-id")
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "yandex.ru",
			},
		},
		{
			name:   "test GET url not found",
			url:    "/aaaaaa",
			method: http.MethodGet,
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name:   "test PUT method not allowed",
			url:    "/",
			method: http.MethodPut,
			want: want{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, repo, cleanup, err := setupTest()
			require.NoError(t, err)
			defer cleanup()

			if test.setup != nil {
				test.setup(repo)
			}

			r := chi.NewRouter()
			r.Use(auth.Auth(testSecretKey))
			r.Method(http.MethodPost, "/", handler.BaseHandler())
			r.Method(http.MethodGet, "/{id}", handler.BaseHandler())
			r.Method(http.MethodPut, "/", handler.BaseHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := newRequestWithUserID(test.method, ts.URL+test.url, test.body)
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

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}

			if test.want.location != "" {
				assert.Equal(t, test.want.location, res.Header.Get("Location"))
			}
		})
	}
}

func TestURLHandler_ShortenHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name    string
		method  string
		body    string
		headers map[string]string
		setup   func(repo repository.URLRepository)
		want    want
	}{
		{
			name:   "test POST success",
			method: http.MethodPost,
			body:   `{"url": "google.ru"}`,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:   "test POST url exists",
			method: http.MethodPost,
			body:   `{"url": "google.ru"}`,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			setup: func(repo repository.URLRepository) {
				ctx := context.WithValue(context.Background(), userIDContextKey, "test-user-id")
				repo.Save(ctx, "abcdef", "google.ru", "test-user-id")
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json",
			},
		},
		{
			name:    "test GET method not allowed",
			method:  http.MethodGet,
			headers: map[string]string{},
			want: want{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "application/json",
			},
		},
		{
			name:   "test POST invalid content type",
			method: http.MethodPost,
			body:   `{"url": "google.ru"}`,
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
			want: want{
				statusCode:  http.StatusUnsupportedMediaType,
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, repo, cleanup, err := setupTest()
			require.NoError(t, err)
			defer cleanup()

			if test.setup != nil {
				test.setup(repo)
			}

			r := chi.NewRouter()
			r.Use(auth.Auth(testSecretKey))
			r.Method(http.MethodPost, "/api/shorten", handler.ShortenHandler())
			r.Method(http.MethodGet, "/api/shorten", handler.ShortenHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := newRequestWithUserID(test.method, ts.URL+"/api/shorten", test.body)
			require.NoError(t, err)

			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			client := &http.Client{}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}

			if res.StatusCode == http.StatusCreated {
				var response model.ResponseShorten
				err := json.NewDecoder(res.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response.Result, "http://localhost:8080/")
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-*.json")
	require.NoError(t, err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	repo, err := repository.NewMemoryStorage(tmpfile.Name())
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), userIDContextKey, "test-user-id")
	repo.Save(ctx, "abcdef", "google.ru", "test-user-id")

	urlService := service.NewURLService(repo, "http://localhost:8080", 100, 5, 30*time.Second, zap.NewNop())
	handler := NewURLHandler(urlService)

	r := chi.NewRouter()
	r.Use(auth.Auth(testSecretKey))
	handlerMiddleware := packgzip.GzipMiddleware(handler.ShortenHandler())
	r.Post("/api/shorten", handlerMiddleware.ServeHTTP)

	ts := httptest.NewServer(r)
	defer ts.Close()

	requestBody := `{"url": "google.ru"}`
	successBody := `{"result": "http://localhost:8080/abcdef"}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req, err := newRequestWithUserID(http.MethodPost, ts.URL+"/api/shorten", "")
		require.NoError(t, err)
		req.Body = io.NopCloser(buf)

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

		assert.Equal(t, http.StatusConflict, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		req, err := newRequestWithUserID(http.MethodPost, ts.URL+"/api/shorten", requestBody)
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
		defer res.Body.Close()

		assert.Equal(t, http.StatusConflict, res.StatusCode)

		zr, err := gzip.NewReader(res.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}

func TestURLHandler_ShortenBatchHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		body   string
		want   int
	}{
		{
			name:   "test POST batch success",
			method: http.MethodPost,
			body: `[
				{"correlation_id": "1", "original_url": "https://google.com  "},
				{"correlation_id": "2", "original_url": "https://yandex.ru  "}
			]`,
			want: http.StatusCreated,
		},
		{
			name:   "test GET method not allowed",
			method: http.MethodGet,
			body:   "",
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "test POST invalid JSON",
			method: http.MethodPost,
			body:   `invalid json`,
			want:   http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, _, cleanup, err := setupTest()
			require.NoError(t, err)
			defer cleanup()

			r := chi.NewRouter()
			r.Use(auth.Auth(testSecretKey))
			r.Method(http.MethodPost, "/api/shorten/batch", handler.ShortenBatchHandler())
			r.Method(http.MethodGet, "/api/shorten/batch", handler.ShortenBatchHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := newRequestWithUserID(test.method, ts.URL+"/api/shorten/batch", test.body)
			require.NoError(t, err)

			if test.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.want, res.StatusCode)

			if res.StatusCode == http.StatusCreated {
				var responses []model.ResponseShortenBatch
				err := json.NewDecoder(res.Body).Decode(&responses)
				require.NoError(t, err)
				assert.Greater(t, len(responses), 0)
			}
		})
	}
}

func TestURLHandler_PingHandler(t *testing.T) {
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
			handler, _, cleanup, err := setupTest()
			require.NoError(t, err)
			defer cleanup()

			r := chi.NewRouter()
			r.Use(auth.Auth(testSecretKey))
			r.Method(http.MethodGet, "/ping", handler.PingHandler())
			r.Method(http.MethodPost, "/ping", handler.PingHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := newRequestWithUserID(test.method, ts.URL+"/ping", "")
			require.NoError(t, err)

			client := &http.Client{}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.want, res.StatusCode)
		})
	}
}

func TestURLHandler_GetAllHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		setup  func(repo repository.URLRepository)
		want   int
	}{
		{
			name:   "test GET all success",
			method: http.MethodGet,
			setup: func(repo repository.URLRepository) {
				ctx := context.WithValue(context.Background(), userIDContextKey, "test-user-id")
				repo.Save(ctx, "abc123", "https://google.com  ", "test-user-id")
				repo.Save(ctx, "def456", "https://yandex.ru  ", "test-user-id")
			},
			want: http.StatusOK,
		},
		{
			name:   "test GET all empty",
			method: http.MethodGet,
			want:   http.StatusNoContent,
		},
		{
			name:   "test POST method not allowed",
			method: http.MethodPost,
			want:   http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, repo, cleanup, err := setupTest()
			require.NoError(t, err)
			defer cleanup()

			if test.setup != nil {
				test.setup(repo)
			}

			r := chi.NewRouter()
			r.Use(auth.Auth(testSecretKey))
			r.Method(http.MethodGet, "/api/user/urls", handler.UserUrlsHandler())
			r.Method(http.MethodPost, "/api/user/urls", handler.UserUrlsHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := newRequestWithUserID(test.method, ts.URL+"/api/user/urls", "")
			require.NoError(t, err)

			client := &http.Client{}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.want, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var responses []model.ResponseGetUserUrls
				err := json.NewDecoder(res.Body).Decode(&responses)
				require.NoError(t, err)

				if test.setup != nil {
					assert.Greater(t, len(responses), 0)
				} else {
					assert.Equal(t, 0, len(responses))
				}
			}
		})
	}
}
