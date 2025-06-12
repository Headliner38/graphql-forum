package graph

import (
	"context"
	"fmt"
	"github.com/headliner38/graphql-forum/graph/storage"
	"sync"

	//"github.com/headliner38/graphql-forum/graph/generated"
	"github.com/graph-gophers/dataloader/v7"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/headliner38/graphql-forum/graph/model"
)

type DepthAndLengthConfig struct { // del
	MaxCommentDepth  int
	MaxCommentLength int
}

type Resolver struct {
	Storage              storage.Repository
	Loaders              map[string]*dataloader.Loader[string, []*model.Comment]
	Cache                *lru.Cache[string, *model.Comment]
	DepthAndLengthConfig DepthAndLengthConfig
	subscriptions        map[string]map[string]chan *model.Comment
	mu                   sync.Mutex // потокобезопасность по ТЗ

}

func (r *Resolver) LoadRepliesForComment(comment *model.Comment) []*model.Comment {
	loader, ok := r.Loaders["CommentReplies"]
	if !ok {
		return nil
	}

	thunk := loader.Load(context.Background(), comment.ID)
	result, err := thunk()
	if err != nil {
		return nil
	}

	return result
	/*var replies []*model.Comment
	for _, c := range r.Comments {
		if c.ParentCommID != nil && *c.ParentCommID == comment.ID {
			replies = append(replies, c)
		}
	}
	return replies*/
}

func NewResolver(cfg DepthAndLengthConfig, storage storage.Repository) (*Resolver, error) { // создание
	cache, err := lru.New[string, *model.Comment](1000)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	resolver := &Resolver{
		Storage:              storage,
		DepthAndLengthConfig: cfg,
		Cache:                cache,
		subscriptions:        make(map[string]map[string]chan *model.Comment),
	}

	resolver.Loaders = NewLoaders(resolver)
	return resolver, nil
}

func (r *Resolver) GetCommentTree(ctx context.Context, postID string, commentID string, maxDepth int) (*model.Comment, error) {
	// Проверяем кэш
	if cached, ok := r.Cache.Get(commentID); ok {
		return cached, nil
	}

	// Загружаем из хранилища
	comment, err := r.Storage.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	if comment.PostID != postID {
		return nil, fmt.Errorf("comment does not belong to specified post")
	}

	fullComment, err := r.loadCommentWithReplies(comment, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to load comment tree: %w", err)
	}

	// Сохраняем в кэш
	r.Cache.Add(commentID, fullComment)
	return fullComment, nil
	/*if cached, ok := r.Cache.Get(commentID); ok {
		return cached, nil
	}

	// Загружаем из хранилища
	comment, err := r.loadCommentWithReplies(r.findCommentByID(commentID), maxDepth)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	r.Cache.Add(commentID, comment)
	return comment, nil*/
}

func (r *Resolver) loadCommentWithReplies(comment *model.Comment, maxDepth int) (*model.Comment, error) { // резолвер для загрузки комментариев со ВСЕМИ ответами (рекурс)
	if maxDepth <= 0 {
		return comment, nil
	}

	replies, err := r.Storage.GetReplies(context.Background(), comment.ID)
	if err != nil {
		return nil, err
	}

	for _, reply := range replies {
		if _, err := r.loadCommentWithReplies(reply, maxDepth-1); err != nil {
			return nil, err
		}
	}

	comment.Replies = replies
	return comment, nil
	/*if maxDepth <= 0 {
		return comment, nil
	}

	loader := r.Loaders["CommentReplies"]
	thunk := loader.Load(context.Background(), comment.ID) // TODO: поправить контекст
	replies, err := thunk()
	if err != nil {
		return nil, err
	}

	for _, reply := range replies {
		_, err := r.loadCommentWithReplies(reply, maxDepth-1)
		if err != nil {
			return nil, err
		}
	}

	comment.Replies = replies
	return comment, nil*/
}

// NotifySubscribers уведомляет подписчиков о новом комментарии
func (r *Resolver) NotifySubscribers(comment *model.Comment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if subscribers, ok := r.subscriptions[comment.PostID]; ok {
		for _, ch := range subscribers {
			select {
			case ch <- comment:
				// Успешно отправлено
			default:
				// Пропускаем если канал полный
			}
		}
	}
}

/*func (r *Resolver) findPostByID(id string) (*model.Post, error) {
	for _, post := range r.Posts {
		if post.ID == id {
			return post, nil
		}
	}
	return nil, fmt.Errorf("пост с id: %s не найден", id)
}*/

/*func (r *Resolver) findCommentByID(id string) *model.Comment { // резолвер для поиска комментария по ID
	for _, comment := range r.Comments {
		if comment.ID == id {
			return comment
		}
	}
	return nil
}*/

/*func (r *Resolver) DebugCheckReplies() { //удалить когда проект будет готов(вывод в терминал для пвет на комментарий к корневому комментарию)
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
}*/
