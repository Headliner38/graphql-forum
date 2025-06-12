package storage

import (
	"context"
	"github.com/headliner38/graphql-forum/graph/model"
)

type Repository interface {
	// Посты
	CreatePost(ctx context.Context, post *model.Post) error
	GetPosts(ctx context.Context) ([]*model.Post, error)
	GetPostByID(ctx context.Context, id string) (*model.Post, error)

	// Комментарии
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	GetCommentByID(ctx context.Context, id string) (*model.Comment, error)
	GetReplies(ctx context.Context, parentID string) ([]*model.Comment, error)
}
