# go-restful-api

Simple example of an api using [golang-standards/project-layout](https://github.com/golang-standards/project-layout) as the project structure, it uses a wrapper of [gin-gonic/gin](https://github.com/gin-gonic/gin) as its HTTP web framework and mysql database for storage.

This example is based on [joaquinlpereyra's](https://github.com/joaquinlpereyra) go projects, most of the /pkg folder wrapper are his owns, so thanks for such a nice work joaquito!

# How to get it running

Get the project
```
go get pclavier92/go-restful-api
```

Create the database 
```
mysql -u root -p < scripts/db_schema.sql
```

Run the API!
```
go run cmd/music/main.go
```

#
### GET Songs

`localhost:3000/songs`


### GET Song by name

`localhost:3000/songs/<name>`

### POST Create song by name

`localhost:3000/songs/<name>`

Body *raw(application/json)*
```
{
	"name": "<name>",
	"duration": "mm:ss",
	"artistId": 1
}
```

### PUT Update song by name

`localhost:3000/songs/<name>`

Body *raw(application/json)*
```
{
	"name": "<name>",
	"duration": "mm:ss",
	"artistId": 1
}
```

### DELETE Delete song by name

`localhost:3000/songs/<name>`

### GET Artists

`localhost:3000/artists`

### GET Artist by name

`localhost:3000/artists/<name>`

### POST Create artist by name

`localhost:3000/artists/<name>`

Body *raw(application/json)*
```
{
	"name": "<name>"
}
```

### DELETE Delete artist by name

`localhost:3000/artists/<name>`
