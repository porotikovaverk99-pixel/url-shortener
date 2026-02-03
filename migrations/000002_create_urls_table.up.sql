-- migrations/000002_url_shortener_table.up.sql
-- Создание таблицы сокращенных URL
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(50) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по short_id
CREATE INDEX idx_urls_short_id ON urls(short_url);

-- Индекс для поиска по original_url
CREATE INDEX idx_urls_original_url ON urls(original_url);

-- Индекс для сортировки по дате создания
CREATE INDEX idx_urls_created_at ON urls(created_at);
