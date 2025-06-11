package graph

import "github.com/headliner38/graphql-forum/graph/model"

// getPaginationParams возвращает значения limit и offset с проверкой(nil)
func getPaginationParams(limit *int32, offset *int32) (int, int) {
	l := 10 // значение по умолчанию
	if limit != nil {
		l = int(*limit)
	}

	o := 0 // значение по умолчанию
	if offset != nil {
		o = int(*offset)
	}

	return l, o
}

// applyPagination применяет пагинацию к слайсу комментариев
func applyPagination(comments []*model.Comment, offset int, limit int) []*model.Comment {
	if offset >= len(comments) {
		return []*model.Comment{}
	}

	end := offset + limit
	if end > len(comments) {
		end = len(comments)
	}

	return comments[offset:end]
}
