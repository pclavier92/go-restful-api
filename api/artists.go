package api

type Artists []Artist

type Artist struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
