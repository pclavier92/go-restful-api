package config

import (
	"os"
	"strings"
)

// H is a simple holder for configuration
type H struct {
	Productive bool
	Job        bool
	Port       string
	Scope      string
	AppVersion string
	DBUser     string
	DBPass     string
	DBHost     string
	DBName     string
}

// New will return a simple holder for our app-wide configuration
func New() H {
	scope, ver := os.Getenv("SCOPE"), os.Getenv("VERSION")
	job := false
	if strings.HasPrefix(scope, "job") {
		scope = "production"
		job = true
	}
	switch scope {
	case "production":
		return H{
			true,
			job,
			"8080",
			"production",
			ver,
			"niceDbUser",
			"niceDBPass",
			"niceDBHost",
			"niceDBName",
		}
	case "test":
		return H{
			false,
			job,
			"8080",
			"test",
			ver,
			"testDbUser",
			"testDBPass",
			"testDBHost",
			"testDBName",
		}
	default:
		return H{
			false,
			job,
			"3000",
			"local",
			"local",
			"root",
			"root",
			"127.0.0.1:3306",
			"Music",
		}
	}
}
