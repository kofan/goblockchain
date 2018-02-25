package appflag

import (
	"flag"

	log "github.com/sirupsen/logrus"
)

type logLevel struct {
	log.Level
}

func (ll *logLevel) String() string {
	return ll.Level.String()
}

func (ll *logLevel) Set(value string) error {
	lvl, err := log.ParseLevel(value)
	if err != nil {
		return err
	}
	ll.Level = lvl
	return nil
}

// LogLevel flag for the command line
func LogLevel(name string, value log.Level, usage string) *log.Level {
	ll := logLevel{value}
	flag.CommandLine.Var(&ll, name, usage)
	return &ll.Level
}
