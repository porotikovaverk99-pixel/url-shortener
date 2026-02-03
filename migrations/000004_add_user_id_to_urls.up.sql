-- migrations/000004_add_user_id_to_urls.up.sql
-- Добавление колонки без внешнего ключа
ALTER TABLE urls ADD COLUMN user_id UUID;

-- Явное создание внешнего ключа с именем и каскадным удалением
ALTER TABLE urls 
ADD CONSTRAINT fk_urls_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Уникальный индекс: один пользователь - один original_url
CREATE UNIQUE INDEX idx_urls_user_original_url 
ON urls(user_id, original_url);

-- Индекс для быстрого поиска URL по пользователю
CREATE INDEX idx_urls_user_id ON urls(user_id);