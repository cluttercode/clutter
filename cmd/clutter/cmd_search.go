package main

import (
	"fmt"
	"strings"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
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

			ent := clutterindex.Entry{Name: name, Attrs: attrs}
			z.Infow("using matcher", "ent", ent)

			matcher, err := ent.Matcher()
			if err != nil {
				return fmt.Errorf("matcher: %w", err)
			}

			src, done, err := ReadIndex(opts.indexPath)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			defer done()

			if err := clutterindex.ForEach(
				src,
				func(ent *clutterindex.Entry) (_ error) {
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
