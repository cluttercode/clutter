package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/internal/pkg/index"
)

func readIndex(c *cli.Context) (*index.Index, error) {
	paths := indexPaths(c)

	z.Debugw("reading index", "paths", paths)

	for _, path := range paths {
		z := z.With("path", path)

		idx, err := readSpecificIndex(path)
		if err != nil {
			if os.IsNotExist(err) {
				z.Warn("file does not exist")
				continue
			}

			if path == "" {
				path = "stdin"
			}

			return nil, fmt.Errorf("%s: %w", path, err)
		}

		z.Info("index read")

		return idx, nil
	}

	return nil, fmt.Errorf("no index file exist")
}

func readSpecificIndex(filename string) (*index.Index, error) {
	if filename == "" {
		return readAdHocIndex()
	}

	return index.ReadFile(filename)
}

func readAdHocIndex() (*index.Index, error) {
	scan, err := scanner.NewScanner(nil, z.Named("scanner"), cfg.Scanner)
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

	return index.NewIndex(ents), nil
}
