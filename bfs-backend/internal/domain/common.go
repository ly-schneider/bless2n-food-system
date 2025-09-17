package domain

type Cents int64

type ListResponse[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}
