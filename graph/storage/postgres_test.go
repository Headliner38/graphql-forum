package storage

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/headliner38/graphql-forum/graph/model"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDBConnStr() string {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	return "postgres://test_user:test_password@" + host + ":5432/postgres_test?sslmode=disable"
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", getTestDBConnStr())
	require.NoError(t, err)

	// Очищаем таблицы перед тестами
	_, err = db.Exec("DROP TABLE IF EXISTS comments, posts CASCADE")
	require.NoError(t, err)

	// Создаем таблицы заново
	_, err = db.Exec(`
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
	`)
	require.NoError(t, err)

	return db
}

func TestPostgresRepository_CreatePost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewPostgresRepository(getTestDBConnStr())
	require.NoError(t, err)
	ctx := context.Background()

	// Тест создания поста
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}

	err = repo.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Проверяем, что пост создан
	posts, err := repo.GetPosts(ctx)
	assert.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, post.ID, posts[0].ID)
	assert.Equal(t, post.Title, posts[0].Title)
	assert.Equal(t, post.Content, posts[0].Content)
	assert.Equal(t, post.Comments, posts[0].Comments)

	// Проверяем получение поста по ID
	retrievedPost, err := repo.GetPostByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post.ID, retrievedPost.ID)
	assert.Equal(t, post.Title, retrievedPost.Title)
	assert.Equal(t, post.Content, retrievedPost.Content)
	assert.Equal(t, post.Comments, retrievedPost.Comments)
}

func TestPostgresRepository_CreateComment(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewPostgresRepository(getTestDBConnStr())
	require.NoError(t, err)
	ctx := context.Background()

	// Создаем пост для комментария
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}
	err = repo.CreatePost(ctx, post)
	require.NoError(t, err)

	// Тест создания комментария
	comment := &model.Comment{
		ID:     "comment1",
		Text:   "Test Comment",
		PostID: post.ID,
	}

	err = repo.CreateComment(ctx, comment)
	assert.NoError(t, err)

	// Проверяем, что комментарий создан
	comments, err := repo.GetCommentsByPostID(ctx, post.ID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, comment.ID, comments[0].ID)
	assert.Equal(t, comment.Text, comments[0].Text)
	assert.Equal(t, comment.PostID, comments[0].PostID)

	// Проверяем получение комментария по ID
	retrievedComment, err := repo.GetCommentByID(ctx, comment.ID)
	assert.NoError(t, err)
	assert.Equal(t, comment.ID, retrievedComment.ID)
	assert.Equal(t, comment.Text, retrievedComment.Text)
	assert.Equal(t, comment.PostID, retrievedComment.PostID)
}

func TestPostgresRepository_GetReplies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewPostgresRepository(getTestDBConnStr())
	require.NoError(t, err)
	ctx := context.Background()

	// Создаем пост
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}
	err = repo.CreatePost(ctx, post)
	require.NoError(t, err)

	// Создаем родительский комментарий
	parentComment := &model.Comment{
		ID:     "parent1",
		Text:   "Parent Comment",
		PostID: post.ID,
	}
	err = repo.CreateComment(ctx, parentComment)
	require.NoError(t, err)

	// Создаем ответы на комментарий
	reply1 := &model.Comment{
		ID:           "reply1",
		Text:         "Reply 1",
		PostID:       post.ID,
		ParentCommID: &parentComment.ID,
	}
	reply2 := &model.Comment{
		ID:           "reply2",
		Text:         "Reply 2",
		PostID:       post.ID,
		ParentCommID: &parentComment.ID,
	}

	err = repo.CreateComment(ctx, reply1)
	require.NoError(t, err)
	err = repo.CreateComment(ctx, reply2)
	require.NoError(t, err)

	// Проверяем получение ответов
	replies, err := repo.GetReplies(ctx, parentComment.ID)
	assert.NoError(t, err)
	assert.Len(t, replies, 2)

	// Проверяем содержимое ответов
	replyIDs := make(map[string]bool)
	for _, reply := range replies {
		replyIDs[reply.ID] = true
	}
	assert.True(t, replyIDs[reply1.ID])
	assert.True(t, replyIDs[reply2.ID])
}

func TestPostgresRepository_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := NewPostgresRepository(getTestDBConnStr())
	require.NoError(t, err)
	ctx := context.Background()

	// Создаем пост
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}
	err = repo.CreatePost(ctx, post)
	require.NoError(t, err)

	// Создаем несколько комментариев
	for i := 0; i < 5; i++ {
		comment := &model.Comment{
			ID:     string(rune('a' + i)),
			Text:   "Comment " + string(rune('a'+i)),
			PostID: post.ID,
		}
		err = repo.CreateComment(ctx, comment)
		require.NoError(t, err)
	}

	// Тест пагинации
	comments, err := repo.GetCommentsByPostID(ctx, post.ID, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	comments, err = repo.GetCommentsByPostID(ctx, post.ID, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	comments, err = repo.GetCommentsByPostID(ctx, post.ID, 2, 4)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
}
