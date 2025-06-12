-- Таблица постов
CREATE TABLE IF NOT EXISTS posts (
                                     id TEXT PRIMARY KEY,
                                     title TEXT NOT NULL,
                                     content TEXT NOT NULL,
                                     comments_enabled BOOLEAN NOT NULL DEFAULT TRUE
);

-- Таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
                                        id TEXT PRIMARY KEY,
                                        text TEXT NOT NULL CHECK (length(text) <= 2000),
    post_id TEXT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    parent_comm_id TEXT REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
    );

-- Индексы для ускорения запросов
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_comm_id);