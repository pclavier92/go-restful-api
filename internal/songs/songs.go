package songs

import (
	"github.com/pclavier92/go-restful-api/api"
	"github.com/pclavier92/go-restful-api/pkg/errors"
	"github.com/pclavier92/go-restful-api/pkg/gin"
	"github.com/pclavier92/go-restful-api/pkg/logs"
	"github.com/pclavier92/go-restful-api/pkg/persist"
)

type persistor interface {
	get(name string) (api.Songs, error)
	create(i api.Song) (bool, error)
	update(i api.Song) (bool, error)
	delete(name string) (bool, error)
}

type db struct {
	persist.Querier
	err errors.Structer
	logs.Printer
}

// Service works as a holder for dependencies of songs
type Service struct {
	db  persistor
	err errors.Structer
	log logs.Printer
}

// API has an HTTP interface for the songs
type API struct {
	s   Service
	err errors.Structer
}

// New will return a new Service for the songs and an API to expose them via HTTP.
func New(sql persist.Querier, log logs.Printer) (*Service, *API) {
	e := errors.Pkg("songs", log)
	s := Service{
		db:  db{sql, e.Struct("db"), log},
		err: e.Struct("service"),
		log: log}
	return &s, &API{s, e.Struct("api")}
}

/*---------------   API   ---------------*/

// GetSongs will retrieve the list of songs.
func (a API) GetSongs(c *gin.Context) (int, interface{}, error) {
	e := a.err.Fn("GetSongs")
	songs, err := a.s.getSongs()
	if err != nil {
		return 500, songs, e.UK(err)
	}
	return 200, songs, nil
}

// GetSongByName will retrive an song by its name
func (a API) GetSongByName(c *gin.Context) (int, interface{}, error) {
	name := c.Param("name")
	e := a.err.Fn("GetSongByName").Tag("name", name)
	song, ok, err := a.s.getSongByName(name)
	if err != nil {
		return 500, nil, e.UK(err)
	} else if !ok {
		return 404, nil, e.NotFound()
	}
	return 200, song, nil
}

// CreateSong will save an Song. Takes a JSON with the new song.
func (a API) CreateSong(c *gin.Context) (int, interface{}, error) {
	e := a.err.Fn("CreateSong")
	var song api.Song
	if err := c.BindJSON(&song); err != nil {
		return 400, nil, e.JSON(err, "binding")
	}
	invalid, err := a.s.saveSong(song)
	if err != nil {
		return 500, nil, e.UK(err)
	}
	if invalid {
		return 400, nil, e.DB(err)
	}
	return 201, nil, nil
}

// UpdateSong will update an Song. Takes a JSON with the updated song.
func (a API) UpdateSong(c *gin.Context) (int, interface{}, error) {
	e := a.err.Fn("UpdateSong")
	var song api.Song
	if err := c.BindJSON(&song); err != nil {
		return 400, nil, e.JSON(err, "binding")
	}
	invalid, err := a.s.saveSong(song)
	if err != nil {
		return 500, nil, e.UK(err)
	}
	if invalid {
		return 400, nil, e.DB(err)
	}
	return 201, nil, nil
}

// DeleteSong will delete an Song. Takes an song's name.
func (a API) DeleteSong(c *gin.Context) (int, interface{}, error) {
	name := c.Param("name")
	e := a.err.Fn("DeleteSong").Tag("name", name)
	invalid, err := a.s.deleteSong(name)
	if err != nil {
		return 500, nil, e.UK(err)
	}
	if invalid {
		return 400, nil, e.DB(err)
	}
	return 201, nil, nil
}

/*--------------- SERVICES ---------------*/

// getSongs will get all songs.
func (s *Service) getSongs() (api.Songs, error) {
	e := s.err.Fn("getSongs")
	songs, err := s.db.get("")
	if err != nil {
		return api.Songs{}, e.Wrap(err, "getting songs from db")
	}
	return songs, nil
}

// getSongByName will get an song by its name.
func (s Service) getSongByName(name string) (api.Song, bool, error) {
	e := s.err.Fn("getSongByName").Tag("name", name)
	songs, err := s.db.get(name)
	if err != nil {
		return api.Song{}, false, e.Wrap(err, "getting contract from db")
	} else if len(songs) != 1 {
		return api.Song{}, false, nil
	}
	i := songs[0]
	return i, true, nil
}

// saveSong will make sure an song is saved in the db
func (s *Service) saveSong(i api.Song) (invalid bool, err error) {
	e := s.err.Fn("saveSong")
	_, ok, err := s.getSongByName(i.Name)
	if err != nil {
		return true, e.Wrap(err, "getting song")
	} else if ok {
		ok, err = s.db.update(i)
	} else {
		ok, err = s.db.create(i)
	}
	if err != nil {
		return true, e.Wrap(err, "saving song")
	} else if !ok {
		return true, nil
	}
	return false, nil
}

// deleteSong will make sure an song is deleted from the db
func (s *Service) deleteSong(name string) (invalid bool, err error) {
	e := s.err.Fn("deleteSong")
	ok, err := s.db.delete(name)
	if err != nil {
		return true, e.Wrap(err, "deleting song")
	} else if !ok {
		return true, nil
	}
	return false, nil
}

/*---------------    DB    ---------------*/

// get will return songs from db
func (db db) get(name string) (api.Songs, error) {
	e := db.err.Fn("get").Tag("name", name)
	var err error
	var rows *persist.Rows
	query := `SELECT id, name, duration, artist_id FROM Songs `
	if name != "" {
		query = query + "WHERE name = ? LIMIT 1"
		rows, err = db.Query(query, name)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, e.Wrap(err, "quering songs from table")
	}
	songs := api.Songs{}
	for rows.Next() {
		i := &api.Song{}
		err := rows.Scan(&(i.Id), &(i.Name), &(i.Duration), &(i.ArtistId))
		if err != nil {
			return nil, e.Wrap(err, "scanning rows")
		}
		songs = append(songs, *i)
	}
	return songs, nil
}

// create will create a new song in the db
func (db db) create(i api.Song) (bool, error) {
	e := db.err.Fn("create")
	query := `INSERT INTO Songs (name, duration, artist_id) VALUES (?, ?, ?)`
	_, err := db.Exec(query, i.Name, i.Duration, i.ArtistId)
	if err != nil {
		return false, e.Wrap(err, "inserting")
	}
	return true, nil
}

// update will update and existing song in the db
func (db db) update(i api.Song) (bool, error) {
	e := db.err.Fn("update")
	query := `UPDATE Songs SET duration = ?, artist_id = ? WHERE name = ?`
	_, err := db.Exec(query, i.Duration, i.ArtistId, i.Name)
	if err != nil {
		return false, e.Wrap(err, "updating")
	}
	return true, nil
}

// delete will delete an existing song from the db
func (db db) delete(name string) (bool, error) {
	e := db.err.Fn("delete")
	query := `DELETE FROM Songs WHERE name = ?`
	_, err := db.Exec(query, name)
	if err != nil {
		return false, e.Wrap(err, "deleting")
	}
	return true, nil
}
