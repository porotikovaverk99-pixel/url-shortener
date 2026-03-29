-- migrations/000001_initial_schema.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица сокращённых URL
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_url VARCHAR(50) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, original_url)  -- уникальность в рамках одного пользователя
);

-- Индексы для производительности
CREATE INDEX idx_urls_short_url ON urls(short_url);
CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE INDEX idx_urls_created_at ON urls(created_at);
CREATE INDEX idx_urls_original_url ON urls(original_url);