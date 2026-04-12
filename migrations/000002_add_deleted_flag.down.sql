-- migrations/000002_add_deleted_flag.down.sql
-- Откат добавления флага удаления

-- Удаляем индекс
DROP INDEX IF EXISTS idx_urls_is_deleted;

-- Удаляем колонку is_deleted
ALTER TABLE urls DROP COLUMN IF EXISTS is_deleted;