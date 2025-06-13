package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/headliner38/graphql-forum/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository - мок для тестирования
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreatePost(ctx context.Context, post *model.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockRepository) GetPosts(ctx context.Context) ([]*model.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Post), args.Error(1)
}

func (m *MockRepository) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockRepository) GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	args := m.Called(ctx, postID, limit, offset)
	return args.Get(0).([]*model.Comment), args.Error(1)
}

func (m *MockRepository) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (m *MockRepository) GetReplies(ctx context.Context, commentID string) ([]*model.Comment, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).([]*model.Comment), args.Error(1)
}

func TestNewResolver(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)

	resolver, err := NewResolver(cfg, mockRepo)
	assert.NoError(t, err)
	assert.NotNil(t, resolver)
	assert.Equal(t, cfg, resolver.DepthAndLengthConfig)
	assert.NotNil(t, resolver.Cache)
	assert.NotNil(t, resolver.Loaders)
}

func TestGetCommentTree(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Тест получения дерева комментариев
	postID := "post1"
	commentID := "comment1"
	comment := &model.Comment{
		ID:           commentID,
		Text:         "Parent comment",
		PostID:       postID,
		ParentCommID: nil,
		Replies:      []*model.Comment{},
	}

	// Настраиваем мок для существующего комментария
	mockRepo.On("GetCommentByID", mock.Anything, commentID).Return(comment, nil)
	mockRepo.On("GetReplies", mock.Anything, commentID).Return([]*model.Comment{}, nil)

	tree, err := resolver.GetCommentTree(context.Background(), postID, commentID, 10)
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	assert.Equal(t, commentID, tree.ID)
	assert.Equal(t, postID, tree.PostID)
	assert.Equal(t, "Parent comment", tree.Text)
	assert.Empty(t, tree.Replies)

	// Тест с несуществующим комментарием
	mockRepo.On("GetCommentByID", mock.Anything, "non_existent_comment").Return(nil, fmt.Errorf("comment not found"))
	_, err = resolver.GetCommentTree(context.Background(), postID, "non_existent_comment", 10)
	assert.Error(t, err)
}

func TestNotifySubscribers(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Создаем тестовый комментарий
	comment := &model.Comment{
		ID:     "comment1",
		Text:   "Test comment",
		PostID: "post1",
	}

	// Создаем канал для подписки
	ch := make(chan *model.Comment, 1)
	resolver.mu.Lock()
	resolver.subscriptions["post1"] = map[string]chan *model.Comment{"sub1": ch}
	resolver.mu.Unlock()

	// Тестируем уведомление
	resolver.NotifySubscribers(comment)

	// Проверяем, что комментарий был отправлен в канал
	received := <-ch
	assert.Equal(t, comment, received)
}
