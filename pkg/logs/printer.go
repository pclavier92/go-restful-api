package logs

import (
	"os"

	"github.com/sirupsen/logrus"
)

type log struct {
	log *logrus.Logger
}

func (l log) Info(title string, args I) {
	fields := logrus.Fields(args)
	l.log.WithFields(fields).Info(title)
}

func (l log) Debug(title string, args I) {
	fields := logrus.Fields(args)
	l.log.WithFields(fields).Debug(title)
}

// I is a simple alias for a map.
type I = map[string]interface{}

// Printer is anything that can print logs
type Printer interface {
	Info(title string, args I)
	Debug(title string, args I)
}

// New returns a logger based on a scope
func New(scope string) (Printer, error) {
	l := logrus.New()
	l.Out = os.Stdout
	if scope == "test" || scope == "" || scope == "local" {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	return log{l}, nil
}
