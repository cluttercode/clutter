package main

import (
	"fmt"
	"strings"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/index"
)

var (
	searchOpts = struct{ regexp, glob bool }{}

	searchCommand = cli.Command{
		Name:    "search",
		Aliases: []string{"s"},
		Usage:   "search index for tags and attributes patterns",
		Flags: []cli.Flag{ // [# search-cli-exp-type-flags #]
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
		},
		Action: func(c *cli.Context) error {
			if searchOpts.glob && searchOpts.regexp {
				return fmt.Errorf("--glob and --regexp are mutually exclusive")
			}

			var (
				name  string
				attrs = map[string]string{}
			)

			if searchOpts.glob {
				attrs["search"] = "glob"
			} else if searchOpts.regexp {
				attrs["search"] = "regexp"
			}

			for _, arg := range c.Args().Slice() {
				parts := strings.SplitN(arg, "=", 2)

				k := parts[0]

				if len(parts) == 1 {
					if name != "" {
						return fmt.Errorf("name already specified")
					}

					name = arg
					continue
				}

				v := parts[1]

				if _, ok := attrs[k]; ok {
					return fmt.Errorf("attribute %q already specified", k)
				}

				attrs[k] = v
			}

			if attrs["search"] == "" {
				attrs["search"] = "exact"
			}

			ent := index.Entry{Name: name, Attrs: attrs}
			z.Infow("using matcher", "ent", ent)

			matcher, err := ent.Matcher()
			if err != nil {
				return fmt.Errorf("matcher: %w", err)
			}

			idx, err := readIndex(c)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			if err := index.ForEach(
				idx,
				func(ent *index.Entry) (_ error) {
					if !matcher(ent) {
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
