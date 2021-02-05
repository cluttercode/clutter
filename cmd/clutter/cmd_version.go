package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

var (
	versionCommand = cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "show current clutter tool version",
		Action: func(*cli.Context) error {
			fmt.Printf("%s %s %s\n", version, commit, date)
			return nil
		},
	}
)
