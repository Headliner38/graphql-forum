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
	if l.r == nil {
		results := make([]*dataloader.Result[[]*model.Comment], len(keys))
		for i := range results {
			results[i] = &dataloader.Result[[]*model.Comment]{
				Error: fmt.Errorf("резолвер не инициализирован"),
			}
		}
		return results
	}

	var results []*dataloader.Result[[]*model.Comment]

	// Предзагрузка всех нужных комментариев за один проход
	repliesMap := make(map[string][]*model.Comment)
	for _, comment := range l.r.Comments {
		if comment.ParentCommID != nil {
			repliesMap[*comment.ParentCommID] = append(repliesMap[*comment.ParentCommID], comment)
		}
	}

	// Формируем результаты в порядке запрошенных ключей
	for _, key := range keys {
		results = append(results, &dataloader.Result[[]*model.Comment]{
			Data: repliesMap[key],
		})
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
