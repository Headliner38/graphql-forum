package graph

import (
	"fmt"

	//"github.com/headliner38/graphql-forum/graph/generated"
	"github.com/headliner38/graphql-forum/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Posts    []*model.Post
	Comments []*model.Comment
}

func (r *Resolver) LoadRepliesForComment(comment *model.Comment) []*model.Comment { // new
	var replies []*model.Comment
	for _, c := range r.Comments {
		if c.ParentCommID != nil && *c.ParentCommID == comment.ID {
			replies = append(replies, c)
		}
	}
	return replies
}

func NewResolver() *Resolver { //need
	return &Resolver{
		Posts:    []*model.Post{},
		Comments: []*model.Comment{},
	}
}

func (r *Resolver) findPostByID(id string) (*model.Post, error) { // need
	for _, post := range r.Posts {
		if post.ID == id {
			return post, nil
		}
	}
	return nil, fmt.Errorf("пост с id: %s не найден", id)
}

func (r *Resolver) DebugCheckReplies() { //new
	fmt.Println("=== DEBUG: Checking replies relationships ===")
	for _, c := range r.Comments {
		if c.ParentCommID != nil {
			fmt.Printf("Comment %s -> Parent %s (Exists: %t)\n",
				c.ID, *c.ParentCommID,
				r.commentExists(*c.ParentCommID, c.PostID))
		}
	}
}

func (r *Resolver) commentExists(id string, postID string) bool { //new
	for _, comment := range r.Comments {
		if comment.ID == id && comment.PostID == postID {
			return true
		}
	}
	return false
}
