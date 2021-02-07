package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/cluttercode/clutter/internal/pkg/resolver"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

var (
	resolveOpts = struct {
		content, loc       string
		prev, next, cyclic bool
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
			&cli.BoolFlag{
				Name:        "cyclic",
				Aliases:     []string{"c"},
				Usage:       "make --next and --prev cyclic",
				Destination: &resolveOpts.cyclic,
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

			var what *clutterindex.Entry

			index, err := clutterindex.Filter(
				src,
				func(ent *clutterindex.Entry) (bool, error) {
					if ent.Loc.Path == loc.Path && ent.Loc.Line == loc.Line && loc.StartColumn >= ent.Loc.StartColumn && loc.EndColumn <= ent.Loc.EndColumn {
						what = ent
					}

					return true, nil
				},
			)

			done()

			if err != nil {
				return fmt.Errorf("index: %w", err)
			}

			if what == nil {
				return fmt.Errorf("no tag at loc")
			}

			z := z.Named("resolver").With("what", what)

			z.Info("resolved tag")

			r := func(z *zap.SugaredLogger, what *clutterindex.Entry, index *clutterindex.Index, _ bool) ([]*clutterindex.Entry, error) {
				return resolver.ResolveList(z, what, index)
			}

			if resolveOpts.next {
				r = resolver.ResolveNext
			} else if resolveOpts.prev {
				r = resolver.ResolvePrev
			}

			ents, err := r(z, what, index, resolveOpts.cyclic)

			if err != nil {
				return fmt.Errorf("resolver: %w", err)
			}

			for _, ent := range ents {
				fmt.Println(ent.String())
			}

			return nil
		},
	}
)
