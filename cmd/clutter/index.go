package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

func readIndex(c *cli.Context) (func() (*clutterindex.Entry, error), func(), error) {
	paths := indexPaths(c)

	z.Debugw("reading index", "paths", paths)

	for _, path := range paths {
		z := z.With("path", path)

		src, done, err := readSpecificIndex(path)
		if err != nil {
			if os.IsNotExist(err) {
				z.Warn("file does not exist")
				continue
			}

			if path == "" {
				path = "stdin"
			}

			return nil, nil, fmt.Errorf("%s: %w", path, err)
		}

		z.Info("index read")

		return src, done, nil
	}

	return nil, nil, fmt.Errorf("no index file exist")
}

func readSpecificIndex(filename string) (src func() (*clutterindex.Entry, error), done func(), err error) {
	if filename == "" {
		scan, err := scanner.NewScanner(z.Named("scanner"), cfg.Scanner)
		if err != nil {
			return nil, nil, fmt.Errorf("new scanner: %w", err)
		}

		elems, err := scan(".", nil)
		if err != nil {
			return nil, nil, fmt.Errorf("scan: %w", err)
		}

		ents, err := parser.ParseElements(elems)
		if err != nil {
			return nil, nil, fmt.Errorf("parser: %w", err)
		}

		src = clutterindex.SliceSource(clutterindex.NewIndex(ents))
		done = func() {}

		return src, done, nil
	} else if src, done, err = clutterindex.FileSource(filename); err != nil {
		return nil, nil, err // do not wrap error.
	}

	return
}
