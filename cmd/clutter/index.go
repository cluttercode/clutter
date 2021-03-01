package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/internal/pkg/index"
)

func hasIndex(c *cli.Context) bool {
	if c.IsSet(indexFlag.Name) {
		return opts.indexPath != ""
	}

	if cfg.UseIndex && opts.indexPath != "" {
		_, err := os.Stat(opts.indexPath)

		return err == nil
	}

	return false
}

func indexPaths(c *cli.Context) []string {
	if c.IsSet(indexFlag.Name) {
		// specifically use name specified.
		return []string{opts.indexPath}
	}

	if cfg.UseIndex {
		// fallback on no index.
		return []string{opts.indexPath, ""}
	}

	// no index
	return []string{""}
}

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

			if path != "" {
				path = path + ": "
			}

			return nil, fmt.Errorf("%s%w", path, err)
		}

		z.Infow("index read", "n", idx.Size())

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

	return index.NewIndex(ents), nil
}

func indexFile(inputPath, actualPath string) (*index.Index, error) {
	elems := make([]*scanner.RawElement, 0, 10)

	if err := scanner.ScanFile(
		z.Named("scan1"),
		cfg.Scanner.Bracket,
		inputPath,
		func(elem *scanner.RawElement) error {
			elem.Loc.Path = actualPath
			elems = append(elems, elem)
			return nil
		},
	); err != nil {
		return nil, err // do not wrap
	}

	ents, err := parser.ParseElements(elems)
	if err != nil {
		return nil, fmt.Errorf("parser: %w", err)
	}

	return index.NewIndex(ents), nil
}
