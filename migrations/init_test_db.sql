-- Подключение к postgres для создания базы данных
\c postgres;

-- Удаляем существующую базу данных и пользователя
DROP DATABASE IF EXISTS postgres_test;
DROP USER IF EXISTS test_user;

-- Создаем пользователя
CREATE USER test_user WITH PASSWORD 'test_password' LOGIN;

-- Создаем базу данных
CREATE DATABASE postgres_test;

-- Подключаемся к тестовой базе данных
\c postgres_test;

-- Создаем таблицы
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    comments_enabled BOOLEAN NOT NULL
);

CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    post_id TEXT NOT NULL REFERENCES posts(id),
    parent_comm_id TEXT REFERENCES comments(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Создание индексов для оптимизации запросов
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_comm_id);

-- Даем права пользователю
GRANT ALL PRIVILEGES ON DATABASE postgres_test TO test_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO test_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO test_user;

-- Передаем владение таблицами posts и comments пользователю test_user
ALTER TABLE posts OWNER TO test_user;
ALTER TABLE comments OWNER TO test_user; 