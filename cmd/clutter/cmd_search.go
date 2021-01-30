package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
	"github.com/cluttercode/clutter/pkg/strmatcher"
)

var (
	searchOpts = struct {
		regexp, glob bool
		attrs        cli.StringSlice
		contextFile  string
	}{}

	searchCommand = cli.Command{
		Name:    "search",
		Aliases: []string{"s"},
		Usage:   "search index for tags and attributes patterns",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "regexp",
				Aliases:     []string{"e", "re"},
				Destination: &searchOpts.regexp,
			},
			&cli.BoolFlag{
				Name:        "glob",
				Aliases:     []string{"g", "gl"},
				Destination: &searchOpts.glob,
			},
			&cli.StringSliceFlag{
				Name:        "attr",
				Aliases:     []string{"a"},
				Destination: &searchOpts.attrs,
			},
		},
		Action: func(c *cli.Context) error {
			if searchOpts.glob && searchOpts.regexp {
				return fmt.Errorf("--glob and --regexp are mutually exclusive")
			}

			src, done, err := ReadIndex(opts.indexPath)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			defer done()

			patternCompiler := strmatcher.CompileExactMatcher

			if searchOpts.regexp {
				patternCompiler = strmatcher.CompileRegexpMatcher
			} else if searchOpts.glob {
				patternCompiler = strmatcher.CompileGlobMatcher
			}

			names, attrs := c.Args().Slice(), searchOpts.attrs.Value()

			filter, err := clutterindex.NewEntriesFilter(patternCompiler, names, attrs)
			if err != nil {
				return fmt.Errorf("matcher: %w", err)
			}

			if err := clutterindex.ForEach(
				src,
				func(ent *clutterindex.Entry) (_ error) {
					if !filter(ent) {
						return
					}

					fmt.Println(ent.String())

					return
				},
			); err != nil {
				return fmt.Errorf("filter: %w", err)
			}

			return nil
		},
	}
)
