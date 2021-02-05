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
		prev, next   bool
	}{}

	resolveCommand = cli.Command{
		Name:    "resolve",
		Aliases: []string{"r"},
		Usage:   "for use by IDEs: resolve tags according to specific instance of a tag",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "prev",
				Aliases:     []string{"p"},
				Usage:       "show only the previous match before the one specified",
				Destination: &resolveOpts.prev,
			},
			&cli.BoolFlag{
				Name:        "next",
				Aliases:     []string{"n"},
				Usage:       "show only the next match after the one specified",
				Destination: &resolveOpts.next,
			},
			&cli.StringFlag{
				Name:        "loc",
				Aliases:     []string{"l"},
				Destination: &resolveOpts.loc,
				Required:    true,
				Usage:       "tag position as path:line.col",
			},
		},
		Action: func(c *cli.Context) error {
			if resolveOpts.next && resolveOpts.prev {
				return fmt.Errorf("--prev and --next are mutually exclusive")
			}

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
				return fmt.Errorf("no tag at loc")
			}

			z.Infow("resolved tag", "tag", given)

			matcher := func(ent *clutterindex.Entry) bool { return given.Name == ent.Name && ent.IsReferredBy(given) }

			if _, search := given.IsSearch(); search {
				if resolveOpts.prev || resolveOpts.next {
					resolveOpts.prev, resolveOpts.next = false, false
					z.Warn("--next and --prev are ignored when resolving a search tag")
				}

				matcher, err = given.Matcher()
				if err != nil {
					return fmt.Errorf("invalid search tag")
				}
			}

			var hold *clutterindex.Entry

			if err := clutterindex.ForEach(
				clutterindex.SliceSource(index),
				func(ent *clutterindex.Entry) error {
					match := matcher(ent)

					z.Debugw("considering", "ent", ent, "match", match)

					if !match {
						return nil
					}

					if resolveOpts.prev {
						if ent.Loc == given.Loc {
							if hold == nil {
								z.Debugw("found given, but nothing held")
								return clutterindex.ErrStop
							}

							z.Debugw("found given, emit held", "ent", hold)

							fmt.Println(hold.String())

							return clutterindex.ErrStop
						}

						hold = ent
						z.Debugw("holding", "ent", hold)

						return nil
					}

					if resolveOpts.next {
						if hold == nil {
							if ent.Loc == given.Loc {
								hold = ent
								z.Debugw("found given", "ent", hold)
							}

							return nil
						}

						z.Debugw("emit current", "ent", hold)

						fmt.Println(ent.String())
						return clutterindex.ErrStop
					}

					fmt.Println(ent.String())

					return nil
				},
			); err != nil {
				return fmt.Errorf("filter: %w", err)
			}

			return nil
		},
	}
)
