package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *HTTPObserver) Send(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := h.client.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil
	}
	return nil
}
