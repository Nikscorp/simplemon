package main

import (
	"github.com/jessevdk/go-flags"
)

func Max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

type Opts struct {
	Config string `long:"config" short:"c" default:"conf/simplemon-conf.yml" description:"Config path"`
}

func parseOpts() (*Opts, error) {
	opts := Opts{}
	p := flags.NewParser(&opts, flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		return nil, err
	}
	return &opts, nil
}
