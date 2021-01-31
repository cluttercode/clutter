package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

const Version = "v0.0.7"

var (
	versionCommand = cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "show current clutter tool version",
		Action: func(*cli.Context) error {
			fmt.Println(Version)
			return nil
		},
	}
)
