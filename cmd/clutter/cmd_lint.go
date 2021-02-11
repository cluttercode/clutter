package main

import (
	"context"
	"fmt"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/index"
	"github.com/cluttercode/clutter/internal/pkg/linter"
)

var (
	lintCommand = cli.Command{
		Name:    "lint",
		Aliases: []string{"l"},
		Usage:   "check tags against lint rules",
		Action: func(c *cli.Context) error {
			linter, err := linter.NewLinter(z.Named("linter"), cfg.Linter)
			if err != nil {
				return fmt.Errorf("linter: %w", err)
			}

			src, done, err := readIndex(c)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			defer done()

			ctx := context.Background()

			pass := true

			if err := index.ForEach(
				src,
				func(ent *index.Entry) error {
					z.Debugw("checking", "loc", ent.Loc)

					failedRulesIndices, err := linter.Lint(ctx, ent)
					if err != nil {
						return err
					}

					z := z.With("loc", ent.Loc)

					if len(failedRulesIndices) != 0 {
						pass = false

						for i, ri := range failedRulesIndices {
							name := linter.Rule(ri).Name
							if name == "" {
								name = fmt.Sprintf("#%d", i)
							}

							fmt.Printf("%v %s\n", ent.Loc, name)
						}
					} else {
						z.Info("entry does not violate any lint rule")
					}

					return nil
				},
			); err != nil {
				return fmt.Errorf("filter: %w", err)
			}

			if !pass {
				return cli.Exit("violations occured", 2)
			}

			return nil
		},
	}
)
