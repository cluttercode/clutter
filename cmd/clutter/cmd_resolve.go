package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/cluttercode/clutter/internal/pkg/index"
	"github.com/cluttercode/clutter/internal/pkg/resolver"
	"github.com/cluttercode/clutter/internal/pkg/scanner"
)

var (
	resolveOpts = struct {
		content, loc       string
		prev, next, cyclic bool
		locFromStdin       bool
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
			&cli.BoolFlag{
				Name:        "loc-from-stdin",
				Destination: &resolveOpts.locFromStdin,
				Usage:       "read file at loc from stdin",
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

			var (
				skipFullIdx bool         // should scan full tree
				idxAtLoc    *index.Index // index built for loc only
				what        *index.Entry // located tag
			)

			filter, err := scanner.NewFilter(z, cfg.Scanner)
			if err != nil {
				return fmt.Errorf("new filter: %w", err)
			}

			if ok, _ := filter(loc.Path, nil); ok && (resolveOpts.locFromStdin || !hasIndex(c)) {
				locPath := loc.Path

				if resolveOpts.locFromStdin {
					locPath = ""
				}

				z.Debugw("preindexing", "path", locPath)

				idxAtLoc, err = indexFile(locPath, loc.Path)
				if err != nil {
					return fmt.Errorf("index loc: %w", err)
				}

				z.Debugw("preindex", "idx", idxAtLoc.Slice())

				if idxAtLoc != nil {
					founds, _ := index.Filter(idxAtLoc, func(ent *index.Entry) (bool, error) {
						return ent.Loc.Contains(*loc), nil
					})

					if founds.Size() == 0 {
						return fmt.Errorf("no tag found at loc")
					}

					if founds.Size() > 1 {
						z.Panicw("found more than single tag at loc", "found", founds)
					}

					what = founds.Slice()[0]

					if what.Attrs["scope"] == loc.Path {
						// optimization: in this case we can skip the full index since we
						// got all data that we need in the file at loc.
						skipFullIdx = true
						z.Debug("loc with matching file scope found at loc, skipping rest of index")
					} else {
						z.Debugw("full index building is required", "tag", what)
					}
				}
			}

			idx := idxAtLoc

			if !skipFullIdx {
				idx1, err := readIndex(c)
				if err != nil {
					return fmt.Errorf("read index: %w", err)
				}

				if idxAtLoc == nil {
					idx = idx1
				} else { // already got data regarding file at loc.

					idx1, _ = index.Filter(idx1, func(ent *index.Entry) (bool, error) {
						// eliminate entries from loc, as we already have them in idxAtLoc.
						return ent.Loc.Path != loc.Path, nil
					})

					idx.Add(idx1.Slice())
				}
			}

			if what == nil {
				_, _ = index.Filter(idx, func(ent *index.Entry) (bool, error) {
					z.Debugw("considering", "loc", ent.Loc)

					if ent.Loc.Contains(*loc) {
						z.Debugw("located", "loc", ent.Loc)

						what = ent

						return false, index.ErrStop
					}

					return false, nil
				})
			}

			if what == nil {
				return fmt.Errorf("no tag at loc")
			}

			z := z.Named("resolver").With("what", what)

			z.Info("resolved tag")

			r := func(z *zap.SugaredLogger, what *index.Entry, idx *index.Index, _ bool) ([]*index.Entry, error) {
				return resolver.ResolveList(z, what, idx)
			}

			if resolveOpts.next {
				r = resolver.ResolveNext
			} else if resolveOpts.prev {
				r = resolver.ResolvePrev
			}

			ents, err := r(z, what, idx, resolveOpts.cyclic)

			if err != nil {
				return fmt.Errorf("resolver: %w", err)
			}

			return index.WriteEntries(os.Stdout, index.NewIndex(ents))
		},
	}
)
