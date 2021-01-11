package main

import (
	"fmt"

	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

func ReadIndex(filename string) (src func() (*clutterindex.Entry, error), done func(), err error) {
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
	} else {
		src, done, err = clutterindex.FileSource(filename)
		if err != nil {
			return nil, nil, fmt.Errorf("index open: %w", err)
		}
	}

	return
}
