-- migrations/000005_fix_original_url_constraint.down.sql
-- Восстанавливаем оригинальный уникальный индекс
DROP INDEX IF EXISTS idx_urls_original_url;

-- Создаем уникальный индекс как было в миграции 000002
CREATE UNIQUE INDEX idx_urls_original_url ON urls(original_url);