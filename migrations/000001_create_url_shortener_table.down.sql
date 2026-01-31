-- migrations/000001_create_urls_table.down.sql
-- Откат создания таблицы URL

-- Удаляем индексы
DROP INDEX IF EXISTS idx_urls_created_at;
DROP INDEX IF EXISTS idx_urls_original_url;
DROP INDEX IF EXISTS idx_urls_short_id;

-- Удаляем таблицу
DROP TABLE IF EXISTS urls;