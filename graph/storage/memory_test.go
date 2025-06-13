package storage

import (
	"context"
	"testing"

	"github.com/headliner38/graphql-forum/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_CreatePost(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Тест создания поста
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}

	err := repo.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Проверяем, что пост создан
	posts, err := repo.GetPosts(ctx)
	assert.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, post, posts[0])

	// Проверяем получение поста по ID
	retrievedPost, err := repo.GetPostByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post, retrievedPost)
}

func TestMemoryRepository_CreateComment(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Создаем пост для комментария
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}
	err := repo.CreatePost(ctx, post)
	assert.NoError(t, err)

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
	assert.Equal(t, comment, comments[0])

	// Проверяем получение комментария по ID
	retrievedComment, err := repo.GetCommentByID(ctx, comment.ID)
	assert.NoError(t, err)
	assert.Equal(t, comment, retrievedComment)
}

func TestMemoryRepository_GetReplies(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Создаем пост
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}
	err := repo.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Создаем родительский комментарий
	parentComment := &model.Comment{
		ID:     "parent1",
		Text:   "Parent Comment",
		PostID: post.ID,
	}
	err = repo.CreateComment(ctx, parentComment)
	assert.NoError(t, err)

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
	assert.NoError(t, err)
	err = repo.CreateComment(ctx, reply2)
	assert.NoError(t, err)

	// Проверяем получение ответов
	replies, err := repo.GetReplies(ctx, parentComment.ID)
	assert.NoError(t, err)
	assert.Len(t, replies, 2)
	assert.Contains(t, replies, reply1)
	assert.Contains(t, replies, reply2)
}

func TestMemoryRepository_Pagination(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Создаем пост
	post := &model.Post{
		ID:       "post1",
		Title:    "Test Post",
		Content:  "Test Content",
		Comments: true,
	}
	err := repo.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Создаем несколько комментариев
	for i := 0; i < 5; i++ {
		comment := &model.Comment{
			ID:     string(rune('a' + i)),
			Text:   "Comment " + string(rune('a'+i)),
			PostID: post.ID,
		}
		err = repo.CreateComment(ctx, comment)
		assert.NoError(t, err)
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
