package queries

import (
	"net/url"
	"strconv"
)

type Paginator struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func InitializePaginator(query url.Values) Paginator {
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil {
		offset = 0
	}
	return Paginator{
		Limit:  limit,
		Offset: offset,
	}
}
