package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/parser"
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
			&cli.StringFlag{
				Name:        "content",
				Aliases:     []string{"c"},
				Usage:       "mark's content",
				Destination: &resolveOpts.content,
			},
		},
		Action: func(c *cli.Context) error {
			loc, err := scanner.ParseLocString(resolveOpts.loc)
			if err != nil {
				return fmt.Errorf("loc: %w", err)
			}

			content := resolveOpts.content
			if content == "" {
				content, err = readMarkAtLoc(*loc)
				if err != nil {
					return err
				}

				z.Infow("located mark", "content", content)
			}

			given, err := parser.ParseElement(&scanner.RawElement{Text: content, Loc: *loc})
			if err != nil {
				return fmt.Errorf("invalid mark: %w", err)
			}

			z.Infow("parsed mark", "entry", given)

			src, done, err := ReadIndex(opts.indexPath)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			defer done()

			if err := clutterindex.ForEach(
				src,
				func(ent *clutterindex.Entry) (_ error) {
					if given.Name != ent.Name {
						return
					}

					if !ent.IsReferredBy(given) {
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

func readMarkAtLoc(loc scanner.Loc) (string, error) {
	re, err := cfg.Scanner.Bracket.Regexp()
	if err != nil {
		return "", fmt.Errorf("invalid bracket: %w", err)
	}

	var (
		text  string
		found bool
	)

	f, err := os.Open(loc.Path)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", loc.Path, err)
	}

	defer f.Close()

	scn := bufio.NewScanner(f)
	for i := 1; !found && scn.Scan(); i++ {
		if i < loc.Line {
			continue
		}

		text = scn.Text()
		found = true
	}

	if err := scn.Err(); err != nil {
		return "", fmt.Errorf("scanner: %w", err)
	}

	if !found {
		return "", fmt.Errorf("not enough lines in file for loc")
	}

	ms := re.FindAllStringIndex(text, -1)

	for _, m := range ms {
		l, r := m[0]+1, m[1]+1

		if loc.Column >= l && loc.Column <= r {
			return strings.TrimSpace(text[l+1 : r-3]), nil
		}
	}

	return "", fmt.Errorf("no mark at loc")
}
