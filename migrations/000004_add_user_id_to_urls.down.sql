-- migrations/000004_add_user_id_to_urls.down.sql
-- Удаление индексов
DROP INDEX IF EXISTS idx_urls_user_original_url;
DROP INDEX IF EXISTS idx_urls_user_id;

-- Удаление колонки user_id
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;