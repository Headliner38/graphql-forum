package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/headliner38/graphql-forum/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePost(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Тест успешного создания поста
	title := "Test Post"
	content := "Test Content"
	comments := true

	mockRepo.On("CreatePost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
		return post.Title == title && post.Content == content && post.Comments == comments
	})).Return(nil)

	mutationResolver := &mutationResolver{resolver}
	post, err := mutationResolver.CreatePost(context.Background(), title, content, comments)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, title, post.Title)
	assert.Equal(t, content, post.Content)
	assert.Equal(t, comments, post.Comments)
}

func TestCreateComment(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Тест успешного создания комментария
	text := "Test Comment"
	postID := "post1"
	parentCommID := "parent1"

	mockRepo.On("CreateComment", mock.Anything, mock.MatchedBy(func(comment *model.Comment) bool {
		return comment.Text == text && comment.PostID == postID && *comment.ParentCommID == parentCommID
	})).Return(nil)

	mutationResolver := &mutationResolver{resolver}
	comment, err := mutationResolver.CreateComment(context.Background(), text, postID, &parentCommID)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, text, comment.Text)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, parentCommID, *comment.ParentCommID)

	// Тест превышения максимальной длины комментария
	longText := string(make([]byte, cfg.MaxCommentLength+1))
	_, err = mutationResolver.CreateComment(context.Background(), longText, postID, nil)
	assert.Error(t, err)
}

func TestPosts(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Тест получения списка постов
	expectedPosts := []*model.Post{
		{
			ID:       "post1",
			Title:    "Test Post 1",
			Content:  "Content 1",
			Comments: true,
		},
		{
			ID:       "post2",
			Title:    "Test Post 2",
			Content:  "Content 2",
			Comments: false,
		},
	}

	mockRepo.On("GetPosts", mock.Anything).Return(expectedPosts, nil)

	queryResolver := &queryResolver{resolver}
	posts, err := queryResolver.Posts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedPosts, posts)
}

func TestComments(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Тест получения комментариев
	postID := "post1"
	limit := int32(10)
	offset := int32(0)

	rootComments := []*model.Comment{
		{
			ID:     "comment1",
			Text:   "Root Comment 1",
			PostID: postID,
		},
	}

	mockRepo.On("GetCommentsByPostID", mock.Anything, postID, int(limit), int(offset)).Return(rootComments, nil)
	mockRepo.On("GetReplies", mock.Anything, "comment1").Return([]*model.Comment{}, nil)

	queryResolver := &queryResolver{resolver}
	comments, err := queryResolver.Comments(context.Background(), postID, limit, offset)

	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Len(t, comments, 1)
	assert.Equal(t, rootComments[0].ID, comments[0].ID)
}

func TestCommentTree(t *testing.T) {
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

	// Настраиваем мок для несуществующего комментария
	mockRepo.On("GetCommentByID", mock.Anything, "non_existent_comment").Return(nil, fmt.Errorf("comment not found"))

	queryResolver := &queryResolver{resolver}
	tree, err := queryResolver.CommentTree(context.Background(), postID, commentID)

	assert.NoError(t, err)
	assert.NotNil(t, tree)
	assert.Equal(t, commentID, tree.ID)
	assert.Equal(t, postID, tree.PostID)
	assert.Equal(t, "Parent comment", tree.Text)
	assert.Empty(t, tree.Replies)

	// Тест с несуществующим комментарием
	_, err = queryResolver.CommentTree(context.Background(), postID, "non_existent_comment")
	assert.Error(t, err)
}

func TestNewCommentSubscription(t *testing.T) {
	cfg := DepthAndLengthConfig{
		MaxCommentDepth:  10,
		MaxCommentLength: 2000,
	}
	mockRepo := new(MockRepository)
	resolver, _ := NewResolver(cfg, mockRepo)

	// Тест подписки на новые комментарии
	postID := "post1"

	// Настраиваем мок для существующего поста
	mockRepo.On("GetPostByID", mock.Anything, postID).Return(&model.Post{
		ID:       postID,
		Title:    "Test Post",
		Content:  "Content",
		Comments: true,
	}, nil)

	// Настраиваем мок для несуществующего поста
	mockRepo.On("GetPostByID", mock.Anything, "non_existent_post").Return(nil, fmt.Errorf("post not found"))

	subscriptionResolver := &subscriptionResolver{resolver}
	ch, err := subscriptionResolver.NewComment(context.Background(), postID)

	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Тест с несуществующим постом
	_, err = subscriptionResolver.NewComment(context.Background(), "non_existent_post")
	assert.Error(t, err)
}
