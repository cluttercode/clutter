package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

var (
	resolveOpts = struct {
		content, loc string
	}{}

	resolveCommand = cli.Command{
		Name:    "resolve",
		Aliases: []string{"r"},
		Usage:   "for use by IDEs: resolve tags according to specific instance of a tag",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "loc",
				Aliases:     []string{"l"},
				Destination: &resolveOpts.loc,
				Required:    true,
				Usage:       "mark position as path:line.col",
			},
		},
		Action: func(c *cli.Context) error {
			loc, err := scanner.ParseLocString(resolveOpts.loc)
			if err != nil {
				return fmt.Errorf("loc: %w", err)
			}

			src, done, err := ReadIndex(opts.indexPath)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			var given *clutterindex.Entry

			index, err := clutterindex.Filter(
				src,
				func(ent *clutterindex.Entry) (bool, error) {
					if ent.Loc.Path == loc.Path && ent.Loc.Line == loc.Line && loc.StartColumn >= ent.Loc.StartColumn && loc.EndColumn <= ent.Loc.EndColumn {
						given = ent
					}

					return true, nil
				},
			)

			done()

			if err != nil {
				return fmt.Errorf("index: %w", err)
			}

			if given == nil {
				return fmt.Errorf("no mark at loc")
			}

			z.Infow("resolved mark", "mark", given)

			matcher := func(ent *clutterindex.Entry) bool { return given.Name == ent.Name && ent.IsReferredBy(given) }

			if _, search := given.IsSearch(); search {
				matcher, err = given.Matcher()
				if err != nil {
					return fmt.Errorf("invalid search mark")
				}
			}

			if err := clutterindex.ForEach(
				clutterindex.SliceSource(index),
				func(ent *clutterindex.Entry) (_ error) {
					if matcher(ent) {
						fmt.Println(ent.String())
					}

					return
				},
			); err != nil {
				return fmt.Errorf("filter: %w", err)
			}

			return nil
		},
	}
)
