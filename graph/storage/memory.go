package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/headliner38/graphql-forum/graph/model"
)

// MemoryRepository реализует Repository для хранения данных в памяти
type MemoryRepository struct {
	posts    []*model.Post
	comments []*model.Comment
	mu       sync.RWMutex // Для потокобезопасности
}

// NewMemoryRepository создает новый экземпляр in-memory хранилища
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		posts:    make([]*model.Post, 0),
		comments: make([]*model.Comment, 0),
	}
}

// CreatePost добавляет новый пост в хранилище
func (r *MemoryRepository) CreatePost(ctx context.Context, post *model.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверка на дубликат ID
	for _, p := range r.posts {
		if p.ID == post.ID {
			return fmt.Errorf("post with ID %s already exists", post.ID)
		}
	}

	r.posts = append(r.posts, post)
	return nil
}

// GetPosts возвращает все посты из хранилища
func (r *MemoryRepository) GetPosts(ctx context.Context) ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем копию, чтобы избежать изменений извне
	posts := make([]*model.Post, len(r.posts))
	copy(posts, r.posts)
	return posts, nil
}

// GetPostByID находит пост по ID
func (r *MemoryRepository) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, post := range r.posts {
		if post.ID == id {
			// Возвращаем копию поста
			return &model.Post{
				ID:       post.ID,
				Title:    post.Title,
				Content:  post.Content,
				Comments: post.Comments,
			}, nil
		}
	}

	return nil, errors.New("post not found")
}

// CreateComment добавляет новый комментарий
func (r *MemoryRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем существование поста
	var postExists bool
	var commentsEnabled bool
	for _, post := range r.posts {
		if post.ID == comment.PostID {
			postExists = true
			commentsEnabled = post.Comments
			break
		}
	}

	if !postExists {
		return errors.New("post not found")
	}

	if !commentsEnabled {
		return errors.New("comments are disabled for this post")
	}

	// Если это ответ, проверяем существование родительского комментария
	if comment.ParentCommID != nil {
		var parentExists bool
		for _, c := range r.comments {
			if c.ID == *comment.ParentCommID && c.PostID == comment.PostID {
				parentExists = true
				break
			}
		}

		if !parentExists {
			return errors.New("parent comment not found")
		}
	}

	// Проверяем дубликат ID
	for _, c := range r.comments {
		if c.ID == comment.ID {
			return fmt.Errorf("comment with ID %s already exists", comment.ID)
		}
	}

	r.comments = append(r.comments, comment)
	return nil
}

// GetCommentsByPostID возвращает корневые комментарии для поста с пагинацией
func (r *MemoryRepository) GetCommentsByPostID(
	ctx context.Context,
	postID string,
	limit, offset int,
) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var rootComments []*model.Comment

	// Собираем корневые комментарии
	for _, comment := range r.comments {
		if comment.PostID == postID && comment.ParentCommID == nil {
			rootComments = append(rootComments, comment)
		}
	}

	// Применяем пагинацию
	if offset >= len(rootComments) {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > len(rootComments) {
		end = len(rootComments)
	}

	// Возвращаем копии комментариев
	result := make([]*model.Comment, end-offset)
	for i := offset; i < end; i++ {
		c := rootComments[i]
		result[i-offset] = &model.Comment{
			ID:           c.ID,
			Text:         c.Text,
			PostID:       c.PostID,
			ParentCommID: c.ParentCommID,
		}
	}

	return result, nil
}

// GetCommentByID находит комментарий по ID
func (r *MemoryRepository) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, comment := range r.comments {
		if comment.ID == id {
			// Возвращаем копию комментария
			return &model.Comment{
				ID:           comment.ID,
				Text:         comment.Text,
				PostID:       comment.PostID,
				ParentCommID: comment.ParentCommID,
			}, nil
		}
	}

	return nil, errors.New("comment not found")
}

// GetReplies возвращает все ответы на указанный комментарий
func (r *MemoryRepository) GetReplies(ctx context.Context, parentID string) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var replies []*model.Comment

	for _, comment := range r.comments {
		if comment.ParentCommID != nil && *comment.ParentCommID == parentID {
			replies = append(replies, comment)
		}
	}

	// Возвращаем копии
	result := make([]*model.Comment, len(replies))
	for i, reply := range replies {
		result[i] = &model.Comment{
			ID:           reply.ID,
			Text:         reply.Text,
			PostID:       reply.PostID,
			ParentCommID: reply.ParentCommID,
		}
	}

	return result, nil
}

// Close реализует метод Close интерфейса (ничего не делает для in-memory)
func (r *MemoryRepository) Close() error {
	return nil
}
