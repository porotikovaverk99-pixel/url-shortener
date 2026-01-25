-- migrations/000002_make_original_url_unique.up.sql
-- Удаление старого индекса и создание уникального индекса для original_url
DROP INDEX IF EXISTS idx_urls_original_url;
CREATE UNIQUE INDEX idx_urls_original_url ON urls(original_url);