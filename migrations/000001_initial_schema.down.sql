-- migrations/000001_initial_schema.down.sql
-- Каскадное удаление в правильном порядке
DROP TABLE IF EXISTS urls CASCADE;
DROP TABLE IF EXISTS users CASCADE;