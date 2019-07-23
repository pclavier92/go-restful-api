package artists

import (
	"github.com/pclavier92/go-restful-api/api"
	"github.com/pclavier92/go-restful-api/pkg/errors"
	"github.com/pclavier92/go-restful-api/pkg/gin"
	"github.com/pclavier92/go-restful-api/pkg/logs"
	"github.com/pclavier92/go-restful-api/pkg/persist"
)

type persistor interface {
	get(name string) (api.Artists, error)
	create(s api.Artist) (bool, error)
	delete(name string) (bool, error)
}

type db struct {
	persist.Querier
	err errors.Structer
	logs.Printer
}

// Service works as a holder for dependencies of pentests
type Service struct {
	db  persistor
	err errors.Structer
	log logs.Printer
}

// API has an HTTP interface for the pentests
type API struct {
	s   Service
	err errors.Structer
}

// New will return a new Service for the pentests and an API to expose them via HTTP.
func New(sql persist.Querier, log logs.Printer) (*Service, *API) {
	e := errors.Pkg("artists", log)
	s := Service{
		db:  db{sql, e.Struct("db"), log},
		err: e.Struct("service"),
		log: log}
	return &s, &API{s, e.Struct("api")}
}

/*---------------   API   ---------------*/

// GetArtists will retrieve the list of artists.
func (a API) GetArtists(c *gin.Context) (int, interface{}, error) {
	e := a.err.Fn("GetArtists")
	artists, err := a.s.getArtists()
	if err != nil {
		return 500, artists, e.UK(err)
	}
	return 200, artists, nil
}

// GetArtistByName will retrive an artist by its name
func (a API) GetArtistByName(c *gin.Context) (int, interface{}, error) {
	name := c.Param("name")
	e := a.err.Fn("GetArtistByName").Tag("name", name)
	artist, ok, err := a.s.getArtistByName(name)
	if err != nil {
		return 500, nil, e.UK(err)
	} else if !ok {
		return 404, nil, e.NotFound()
	}
	return 200, artist, nil
}

// CreateArtist will save an Artist. Takes a JSON with the new artist.
func (a API) CreateArtist(c *gin.Context) (int, interface{}, error) {
	e := a.err.Fn("CreateArtist")
	var artist api.Artist
	if err := c.BindJSON(&artist); err != nil {
		return 400, nil, e.JSON(err, "binding")
	}
	invalid, err := a.s.saveArtist(artist)
	if err != nil {
		return 500, nil, e.UK(err)
	}
	if invalid {
		return 400, nil, e.DB(err)
	}
	return 201, nil, nil
}

// DeleteArtist will delete an Artist. Takes an artist's name.
func (a API) DeleteArtist(c *gin.Context) (int, interface{}, error) {
	name := c.Param("name")
	e := a.err.Fn("DeleteArtist").Tag("name", name)
	invalid, err := a.s.deleteArtist(name)
	if err != nil {
		return 500, nil, e.UK(err)
	}
	if invalid {
		return 400, nil, e.DB(err)
	}
	return 201, nil, nil
}

/*--------------- SERVICES ---------------*/

// getArtists will get all artists.
func (s *Service) getArtists() (api.Artists, error) {
	e := s.err.Fn("getArtists")
	artists, err := s.db.get("")
	if err != nil {
		return api.Artists{}, e.Wrap(err, "getting artists from db")
	}
	return artists, nil
}

// getArtistByName will get an artist by its name.
func (s Service) getArtistByName(name string) (api.Artist, bool, error) {
	e := s.err.Fn("getArtistByName").Tag("name", name)
	artists, err := s.db.get(name)
	if err != nil {
		return api.Artist{}, false, e.Wrap(err, "getting contract from db")
	} else if len(artists) != 1 {
		return api.Artist{}, false, nil
	}
	artist := artists[0]
	return artist, true, nil
}

// saveArtist will make sure an artist is saved in the db
func (s *Service) saveArtist(i api.Artist) (invalid bool, err error) {
	e := s.err.Fn("saveArtist")
	ok, err := s.db.create(i)
	if err != nil {
		return true, e.Wrap(err, "saving artist")
	} else if !ok {
		return true, nil
	}
	return false, nil
}

// deleteArtist will make sure an artist is deleted from the db
func (s *Service) deleteArtist(name string) (invalid bool, err error) {
	e := s.err.Fn("deleteArtist")
	ok, err := s.db.delete(name)
	if err != nil {
		return true, e.Wrap(err, "deleting artist")
	} else if !ok {
		return true, nil
	}
	return false, nil
}

/*---------------    DB    ---------------*/

// get will return artists from db
func (db db) get(name string) (api.Artists, error) {
	e := db.err.Fn("get").Tag("name", name)
	var err error
	var rows *persist.Rows
	query := `SELECT id, name FROM Artists `
	if name != "" {
		query = query + "WHERE name = ? LIMIT 1"
		rows, err = db.Query(query, name)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, e.Wrap(err, "quering artists from table")
	}
	artists := api.Artists{}
	for rows.Next() {
		s := &api.Artist{}
		err := rows.Scan(&(s.Id), &(s.Name))
		if err != nil {
			return nil, e.Wrap(err, "scanning rows")
		}
		artists = append(artists, *s)
	}
	return artists, nil
}

// create will create a new artist in the db
func (db db) create(s api.Artist) (bool, error) {
	e := db.err.Fn("create")
	query := `INSERT INTO Artists (name) VALUES (?)`
	_, err := db.Exec(query, s.Name)
	if err != nil {
		return false, e.Wrap(err, "inserting")
	}
	return true, nil
}

// delete will delete an existing artist from the db
func (db db) delete(name string) (bool, error) {
	e := db.err.Fn("delete")
	query := `DELETE FROM Artists WHERE name = ?`
	_, err := db.Exec(query, name)
	if err != nil {
		return false, e.Wrap(err, "deleting")
	}
	return true, nil
}
