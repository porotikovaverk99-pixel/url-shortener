package model

// RequestShorten представляет запрос на сокращение одного URL.
type RequestShorten struct {
	URL string `json:"url"` // Оригинальный URL для сокращения
}

// RequestShortenBatch представляет запрос на пакетное сокращение URL.
type RequestShortenBatch struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для связи с ответом
	OriginalURL   string `json:"original_url"`   // Оригинальный URL для сокращения
}

// ResponseShorten представляет ответ на сокращение одного URL.
type ResponseShorten struct {
	Result string `json:"result"` // Сокращенный URL
}

// ResponseShortenBatch представляет ответ на пакетное сокращение URL.
type ResponseShortenBatch struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор из запроса
	ShortURL      string `json:"short_url"`      // Сокращенный URL
}

// ResponseGetUserUrls представляет ответ на запрос списка URL пользователя.
type ResponseGetUserUrls struct {
	ShortURL    string `json:"short_url"`    // Сокращенный URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
}

// BatchResult содержит результат пакетной обработки.
type BatchResult struct {
	Responses  []ResponseShortenBatch // Список ответов на пакетный запрос
	CreatedNew bool                   // Были ли созданы новые записи
}

// BatchItem представляет элемент для пакетного сохранения.
type BatchItem struct {
	ShortURL    string // Сгенерированный короткий URL
	OriginalURL string // Оригинальный URL
}

// DeleteTask представляет задачу на удаление URL.
// generate:reset
type DeleteTask struct {
	UserID    string   // Идентификатор пользователя
	ShortURLs []string // Список коротких URL для удаления
}
