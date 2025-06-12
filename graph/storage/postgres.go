package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/headliner38/graphql-forum/graph/model"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// CreatePost создает новый пост в базе данных
func (r *PostgresRepository) CreatePost(ctx context.Context, post *model.Post) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO posts(id, title, content, comments_enabled) VALUES($1, $2, $3, $4)",
		post.ID, post.Title, post.Content, post.Comments,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("post with this ID already exists")
		}
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

// GetPosts возвращает все посты из базы данных
func (r *PostgresRepository) GetPosts(ctx context.Context) ([]*model.Post, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, content, comments_enabled FROM posts ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Comments); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return posts, nil
}

// GetPostByID возвращает пост по его ID
func (r *PostgresRepository) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRowContext(ctx,
		"SELECT id, title, content, comments_enabled FROM posts WHERE id = $1", id).
		Scan(&post.ID, &post.Title, &post.Content, &post.Comments)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

// CreateComment создает новый комментарий
func (r *PostgresRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, существует ли пост
	var commentsEnabled bool
	err = tx.QueryRowContext(ctx,
		"SELECT comments_enabled FROM posts WHERE id = $1", comment.PostID).
		Scan(&commentsEnabled)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("post not found")
		}
		return fmt.Errorf("failed to check post: %w", err)
	}

	if !commentsEnabled {
		return fmt.Errorf("comments are disabled for this post")
	}

	// Если это ответ на другой комментарий, проверяем его существование
	if comment.ParentCommID != nil {
		var exists bool
		err = tx.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND post_id = $2)",
			*comment.ParentCommID, comment.PostID).
			Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check parent comment: %w", err)
		}

		if !exists {
			return fmt.Errorf("parent comment not found")
		}
	}

	// Создаем комментарий
	_, err = tx.ExecContext(ctx,
		`INSERT INTO comments(id, text, post_id, parent_comm_id) 
		VALUES($1, $2, $3, $4)`,
		comment.ID, comment.Text, comment.PostID, comment.ParentCommID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("comment with this ID already exists")
		}
		return fmt.Errorf("failed to create comment: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCommentsByPostID возвращает корневые комментарии для поста с пагинацией
func (r *PostgresRepository) GetCommentsByPostID(
	ctx context.Context,
	postID string,
	limit, offset int,
) ([]*model.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, text, post_id, parent_comm_id 
		FROM comments 
		WHERE post_id = $1 AND parent_comm_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		postID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.Text, &c.PostID, &c.ParentCommID); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return comments, nil
}

// GetCommentByID возвращает комментарий по его ID
func (r *PostgresRepository) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.QueryRowContext(ctx,
		`SELECT id, text, post_id, parent_comm_id 
		FROM comments 
		WHERE id = $1`, id).
		Scan(&comment.ID, &comment.Text, &comment.PostID, &comment.ParentCommID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// GetReplies возвращает все ответы на указанный комментарий
func (r *PostgresRepository) GetReplies(ctx context.Context, parentID string) ([]*model.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, text, post_id, parent_comm_id 
		FROM comments 
		WHERE parent_comm_id = $1
		ORDER BY created_at ASC`,
		parentID)

	if err != nil {
		return nil, fmt.Errorf("failed to query replies: %w", err)
	}
	defer rows.Close()

	var replies []*model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.Text, &c.PostID, &c.ParentCommID); err != nil {
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}
		replies = append(replies, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return replies, nil
}

// Close закрывает соединение с базой данных
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
