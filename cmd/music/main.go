package main

import (
	"github.com/pclavier92/go-restful-api/config"
	"github.com/pclavier92/go-restful-api/internal/artists"
	"github.com/pclavier92/go-restful-api/internal/songs"
	"github.com/pclavier92/go-restful-api/pkg/gin"
	"github.com/pclavier92/go-restful-api/pkg/logs"
	"github.com/pclavier92/go-restful-api/pkg/persist"
	// "github.com/pclavier92/go-restful-api/intenal/users"
)

func main() {
	cfg := config.New()
	log, err := logs.New(cfg.Scope)
	if err != nil {
		panic(err)
	}
	db, err := persist.New(cfg, log)
	if err != nil {
		panic(err)
	}
	log.Info("Starting up API", logs.I{"scope": cfg.Scope})
	e := gin.New(cfg.Port)
	_, songsAPI := songs.New(db, log)
	_, artistsAPI := artists.New(db, log)
	//_, usersAPI := users.New(db, log)

	e.UseLogger()
	{
		i := e.Group("/songs")
		{
			i.GET("", songsAPI.GetSongs)
			i.GET("/:name", songsAPI.GetSongByName)
			i.POST("/:name", songsAPI.CreateSong)
			i.PUT("/:name", songsAPI.UpdateSong)
			i.DELETE("/:name", songsAPI.DeleteSong)
		}
		s := e.Group("/artists")
		{
			s.GET("", artistsAPI.GetArtists)
			s.GET("/:name", artistsAPI.GetArtistByName)
			s.POST("/:name", artistsAPI.CreateArtist)
			s.DELETE("/:name", artistsAPI.DeleteArtist)
		}
		// u := e.Group("/user")
		// {
		// 	u.GET("/:id", usersAPI.GetUserById)
		//  u.POST("/", usersAPI.CreateUser)
		// 	u.DELETE("/:id", usersAPI.DeleteUser)
		// }
	}
	err = e.Run()
	if err != nil {
		panic(err)
	}
}
