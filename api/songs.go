package api

type Songs []Song

type Song struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
	ArtistId int    `json:"artistId"`
}
