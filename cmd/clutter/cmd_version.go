package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

const Version = "v0.0.2"

var (
	versionCommand = cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Action: func(*cli.Context) error {
			fmt.Println(Version)
			return nil
		},
	}
)
