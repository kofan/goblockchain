package appflag

import (
	"errors"
	"flag"
	"strconv"
)

type difficulty struct {
	uint8
}

func (df *difficulty) String() string {
	return strconv.FormatUint(uint64(df.uint8), 10)
}

func (df *difficulty) Set(value string) error {
	d, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return err
	}
	if d > 255 {
		return errors.New("Difficulty cannot be bigger than 255")
	}
	df.uint8 = uint8(d)
	return nil
}

// Difficulty flag for the command line
func Difficulty(name string, value uint8, usage string) *uint8 {
	df := difficulty{value}
	flag.CommandLine.Var(&df, name, usage)
	return &df.uint8
}
