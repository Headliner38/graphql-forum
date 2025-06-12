package graph

import (
	"context"
	"fmt"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/headliner38/graphql-forum/graph/model"
)

type CommentLoader struct {
	r *Resolver
}

func (l *CommentLoader) LoadReplies(ctx context.Context, keys []string) []*dataloader.Result[[]*model.Comment] {
	// сначала проверка на инициализацию, потом начинается уже реализация после следующего комментария
	if l.r == nil || l.r.Storage == nil {
		results := make([]*dataloader.Result[[]*model.Comment], len(keys))
		for i := range results {
			results[i] = &dataloader.Result[[]*model.Comment]{
				Error: fmt.Errorf("резолвер или хранилище не инициализированы"),
			}
		}
		return results
	}

	results := make([]*dataloader.Result[[]*model.Comment], len(keys))

	for i, key := range keys {
		replies, err := l.r.Storage.GetReplies(ctx, key)
		if err != nil {
			results[i] = &dataloader.Result[[]*model.Comment]{
				Error: fmt.Errorf("ошибка загрузки ответов: %w", err),
			}
			continue
		}

		copiedReplies := make([]*model.Comment, len(replies)) // копируем, чтобы не изменились извне
		for j, reply := range replies {
			copiedReplies[j] = &model.Comment{
				ID:           reply.ID,
				Text:         reply.Text,
				PostID:       reply.PostID,
				ParentCommID: reply.ParentCommID,
			}
		}
		results[i] = &dataloader.Result[[]*model.Comment]{
			Data: copiedReplies,
		}
	}
	return results
}

func NewLoaders(r *Resolver) map[string]*dataloader.Loader[string, []*model.Comment] {
	commentLoader := &CommentLoader{r: r}
	loaders := make(map[string]*dataloader.Loader[string, []*model.Comment])
	loaders["CommentReplies"] = dataloader.NewBatchedLoader[string, []*model.Comment](
		commentLoader.LoadReplies,
		dataloader.WithCache[string, []*model.Comment](&dataloader.NoCache[string, []*model.Comment]{}),
	)
	return loaders
}
