package model

type RequestShorten struct {
	URL string `json:"url"`
}

type RequestShortenBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseShorten struct {
	Result string `json:"result"`
}

type ResponseShortenBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ResponseGetUserUrls struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchResult struct {
	Responses  []ResponseShortenBatch
	CreatedNew bool
}

type BatchItem struct {
	ShortURL    string
	OriginalURL string
}
