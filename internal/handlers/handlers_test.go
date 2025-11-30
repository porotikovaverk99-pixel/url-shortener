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
)

func TestURLHandler(t *testing.T) {
	type want struct {
		statusCode int
		response string
		contentType string
		location string
	}
	stor := storage.NewMemoryStorage()
	stor.Save("abcdef", "yandex.ru")
	tests := []struct {
		name string
		url string
		method string
		body io.Reader
		storage storage.URLStorage
		want want
	} {
		{
			name: "test POST",
			url: "/",
			method: http.MethodPost,
			body: strings.NewReader("yandex.ru"),
			storage: stor,
			want: want {
				statusCode: 201,
				contentType: "",
				location: "",
			},
		},
		{
			name: "test POST no data",
			url: "/",
			method: http.MethodPost,
			body: strings.NewReader(""),
			storage: stor,
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
			storage: stor,
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
			storage: stor,
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
		{
			name: "test GET no id",
			url: "/",
			method: http.MethodGet,
			body: nil,
			storage: stor,
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
			storage: stor,
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.url, test.body)
			w := httptest.NewRecorder()
			handler := URLHandler(test.storage)
			handler(w, request)
			res := w.Result()
			assert.Equal(t, test.want.statusCode, res.StatusCode)
			if request.Method == http.MethodPost && res.StatusCode == 201 {
				defer res.Body.Close() 
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				responseURL := string(body)
				parts := strings.Split(responseURL, "/")
				short := parts[len(parts)-1]   
				original, err := test.storage.Get(short)
				assert.NoError(t, err)
				assert.Equal(t, "yandex.ru", original)
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}