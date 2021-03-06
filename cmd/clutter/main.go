package main

import (
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		os.Exit(1)
	}
}
