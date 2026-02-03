-- migrations/000004_add_user_id_to_urls.down.sql
-- Удаление внешнего ключа по имени
ALTER TABLE urls DROP CONSTRAINT IF EXISTS fk_urls_user_id;

-- Удаление индексов
DROP INDEX IF EXISTS idx_urls_user_original_url;
DROP INDEX IF EXISTS idx_urls_user_id;

-- Удаление колонки
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;