-- migrations/000002_add_deleted_flag.up.sql
ALTER TABLE urls 
ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;

-- Индекс для быстрого поиска удаленных URL
CREATE INDEX idx_urls_is_deleted ON urls(is_deleted);