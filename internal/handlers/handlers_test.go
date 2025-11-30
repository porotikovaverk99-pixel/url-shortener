package handlers

import (
	"io"
	"strings"
	"testing"
	"net/http"
	"net/httptest"
	"github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
	"url-shortener/internal/storage"
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
			method: http.MethodPOST,
			body: strings.NewReader("yandex.ru"),
			storage: stor,
			want: want {
				statusCode: 201,
				contentType: "text/plain; charset=utf-8",
				location: ""
			},
		},
		{
			name: "test POST no data",
			url: "/",
			method: http.MethodPOST,
			body: strings.NewReader(""),
			storage: stor,
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: ""
			},
		},
		{
			name: "test GET",
			url: "/abcdef",
			method: http.MethodGET,
			body: nil,
			storage: stor,
			want: want {
				statusCode: 307,
				contentType: "text/plain; charset=utf-8",
				location: "yandex.ru"
			},
		},
		{
			name: "test GET no url",
			url: "/aaaaaa",
			method: http.MethodGET,
			body: nil,
			storage: stor,
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: ""
			},
		},
		{
			name: "test GET no id",
			url: "/",
			method: http.MethodGET,
			body: nil,
			storage: stor,
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: ""
			},
		},
		{
			name: "test PUT",
			url: "/",
			method: http.MethodPUT,
			body: nil,
			storage: stor,
			want: want {
				statusCode: 400,
				contentType: "text/plain; charset=utf-8",
				location: ""
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
			if request.Method == http.MethodPOST && res.StatusCode == 201 {
				defer res.Body.Close() 
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				short := strings.TrimPrefix(body, "/")
				original, err := test.storage.Get(short)
				assert.NoError(t, err)
				assert.Equal(t, string(test.body), original)
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}