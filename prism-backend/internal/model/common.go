package model

type PaginationParams struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Sort  string `query:"sort"`
	Order string `query:"order"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListResponse[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type DataResponse[T any] struct {
	Data T `json:"data"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
