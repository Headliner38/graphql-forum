package graph

import (
	"fmt"

	//"github.com/headliner38/graphql-forum/graph/generated"
	"github.com/graph-gophers/dataloader/v7"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/headliner38/graphql-forum/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type DepthAndLengthConfig struct { // del
	MaxCommentDepth  int
	MaxCommentLength int
}

type Resolver struct {
	Posts                []*model.Post
	Comments             []*model.Comment
	Loaders              map[string]*dataloader.Loader[string, []*model.Comment]
	Cache                *lru.Cache[string, *model.Comment]
	DepthAndLengthConfig DepthAndLengthConfig //MaxDepth int
}

func (r *Resolver) LoadRepliesForComment(comment *model.Comment) []*model.Comment { // резолвер для подгрузки ответов на комментарий
	var replies []*model.Comment
	for _, c := range r.Comments {
		if c.ParentCommID != nil && *c.ParentCommID == comment.ID {
			replies = append(replies, c)
		}
	}
	return replies
}

func NewResolver(cfg DepthAndLengthConfig) (*Resolver, error) { // создание
	cache, err := lru.New[string, *model.Comment](1000)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	resolver := &Resolver{
		Posts:                []*model.Post{},
		Comments:             []*model.Comment{},
		DepthAndLengthConfig: cfg,
		Cache:                cache,
	}

	resolver.Loaders = NewLoaders(resolver)
	return resolver, nil
}

func (r *Resolver) findPostByID(id string) (*model.Post, error) { // резолвер для поиска поста по ID, для того, чтобы оставлять комментарии
	for _, post := range r.Posts {
		if post.ID == id {
			return post, nil
		}
	}
	return nil, fmt.Errorf("пост с id: %s не найден", id)
}

func (r *Resolver) findCommentByID(id string) *model.Comment { // резолвер для поиска комментария по ID
	for _, comment := range r.Comments {
		if comment.ID == id {
			return comment
		}
	}
	return nil
}

func (r *Resolver) DebugCheckReplies() { //удалить когда проект будет готов(вывод в терминал для проверки, привязывается ли ответ на комментарий к корневому комментарию)
	fmt.Println("=== DEBUG: Checking replies relationships ===")
	for _, c := range r.Comments {
		if c.ParentCommID != nil {
			fmt.Printf("Comment %s -> Parent %s (Exists: %t)\n",
				c.ID, *c.ParentCommID,
				r.commentExists(*c.ParentCommID, c.PostID))
		}
	}
}

func (r *Resolver) commentExists(id string, postID string) bool { //удалить когда проект будет готов(проверка на существование комментария)
	for _, comment := range r.Comments {
		if comment.ID == id && comment.PostID == postID {
			return true
		}
	}
	return false
}
