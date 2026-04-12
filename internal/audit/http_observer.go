package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	resp, err := h.client.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("audit server returned %s: %s", resp.Status, string(body))
}
