package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/internal/pkg/index"
)

func readIndex(c *cli.Context) (func() (*index.Entry, error), func(), error) {
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

func readAdHocIndex() (func() (*index.Entry, error), error) {
	scan, err := scanner.NewScanner(z.Named("scanner"), cfg.Scanner)
	if err != nil {
		return nil, fmt.Errorf("new scanner: %w", err)
	}

	elems, err := scan(".", nil)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	ents, err := parser.ParseElements(elems)
	if err != nil {
		return nil, fmt.Errorf("parser: %w", err)
	}

	return index.SliceSource(index.NewIndex(ents)), nil
}

func readSpecificIndex(filename string) (src func() (*index.Entry, error), done func(), err error) {
	if filename == "" {
		done = func() {}
		src, err = readAdHocIndex()
	} else {
		src, done, err = index.FileSource(filename)
	}

	return
}
