-- migrations/000005_fix_original_url_constraint.up.sql
-- Удаляем старый уникальный индекс на original_url
DROP INDEX IF EXISTS idx_urls_original_url;

-- Создаем обычный (не уникальный) индекс для поиска
CREATE INDEX idx_urls_original_url ON urls(original_url);

