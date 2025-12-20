package handlers

import (
	"io"
	"strings"
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"context"
)

func TestURLHandler(t *testing.T) {
	type want struct {
		statusCode int
		response string
		contentType string
		location string
	}
	
	tests := []struct {
		name string
		url string
		method string
		body io.Reader
		storage storage.URLStorage
		baseURL string
		want want
	} {
		{
			name: "test POST",
			url: "/",
			method: http.MethodPost,
			body: strings.NewReader("google.ru"),
			storage: storage.NewMemoryStorage(),
			baseURL: "http://localhost:8080",
			want: want {
				statusCode: 201,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
		{
			name: "test POST no data",
			url: "/",
			method: http.MethodPost,
			body: strings.NewReader(""),
			storage: storage.NewMemoryStorage(),
			baseURL: "http://localhost:8080",
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
		{
			name: "test GET",
			url: "/abcdef",
			method: http.MethodGet,
			body: nil,
			storage: func() storage.URLStorage {
				stor := storage.NewMemoryStorage()
				stor.Save(context.Background(), "abcdef", "yandex.ru")
				return stor
			}(),
			baseURL: "http://localhost:8080",
			want: want {
				statusCode: 307,
				contentType: "",
				location: "yandex.ru",
			},
		},
		{
			name: "test GET no url",
			url: "/aaaaaa",
			method: http.MethodGet,
			body: nil,
			storage: storage.NewMemoryStorage(),
			baseURL: "http://localhost:8080",
			want: want {
				statusCode: 404,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
		{
			name: "test GET no id",
			url: "/",
			method: http.MethodGet,
			body: nil,
			storage: storage.NewMemoryStorage(),
			baseURL: "http://localhost:8080",
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
		{
			name: "test PUT",
			url: "/",
			method: http.MethodPut,
			body: nil,
			storage: storage.NewMemoryStorage(),
			baseURL: "http://localhost:8080",
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := chi.NewRouter()
			handler := URLHandler(test.storage, test.baseURL)
			r.Post("/", handler)
			r.Get("/{id}", handler)
			r.HandleFunc("/*", handler)

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(test.method, ts.URL + test.url, test.body)
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
			if test.method == http.MethodPost && res.StatusCode == 201 {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				responseURL := string(body)
				parts := strings.Split(responseURL, "/")
				short := parts[len(parts)-1]   
				original, err := test.storage.Get(context.Background(), short)
				assert.NoError(t, err)
				assert.Equal(t, "google.ru", original)
			} 
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}