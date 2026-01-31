-- migrations/000002_make_original_url_unique.down.sql
-- Восстановление обычного индекса
DROP INDEX IF EXISTS idx_urls_original_url;
CREATE INDEX idx_urls_original_url ON urls(original_url);