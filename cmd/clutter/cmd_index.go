package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/clutter/clutterindex"
)

var (
	indexOpts = struct {
		watch     bool
		interval  time.Duration
		noINotify bool
	}{
		interval: 30 * time.Second,
	}

	indexCommand = cli.Command{
		Name: "index",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "watch",
				Aliases:     []string{"w"},
				Destination: &indexOpts.watch,
			},
			&cli.DurationFlag{
				Name:        "poll-interval",
				Aliases:     []string{"pi", "ival"},
				Value:       indexOpts.interval,
				Destination: &indexOpts.interval,
			},
			&cli.BoolFlag{
				Name:        "no-inotify",
				Aliases:     []string{"nin"},
				Destination: &indexOpts.noINotify,
			},
		},
		Aliases: []string{"i"},
		Usage:   "generate index database",
		Action: func(c *cli.Context) error {
			roots := c.Args().Slice()

			if len(roots) == 0 {
				roots = []string{"."}
			}

			scan := func() error {
				z.Info("scanning")

				scan, err := scanner.NewScanner(z.Named("scanner"), cfg.Scanner)
				if err != nil {
					return fmt.Errorf("new scanner: %w", err)
				}

				var elems []*scanner.RawElement

				for _, root := range roots {
					elems1, err := scan(root, func(e *scanner.RawElement) error {
						z.Infow("found", "element", e)
						return nil
					})

					if err != nil {
						return fmt.Errorf("%q scan: %w", root, err)
					}

					elems = append(elems, elems1...)
				}

				ents, err := parser.ParseElements(elems)
				if err != nil {
					return fmt.Errorf("parser: %w", err)
				}

				index := clutterindex.NewIndex(ents)

				z.Infow("writing index", "n", index.Size())

				if err := clutterindex.Write(opts.indexPath, index); err != nil {
					return fmt.Errorf("index write: %w", err)
				}

				return nil
			}

			errRefresh := fmt.Errorf("refresh")

			for {
				if err := scan(); err != nil {
					return fmt.Errorf("scan: %w", err)
				}

				if !indexOpts.watch {
					return nil
				}

				watcher, err := fsnotify.NewWatcher()
				if err != nil {
					return fmt.Errorf("watcher: %w", err)
				}

				filter, err := scanner.NewFilter(z, cfg.Scanner)
				if err != nil {
					z.Panicw("new filter error", "err", err)
				}

				if err := filepath.Walk(".", func(path string, fi os.FileInfo, err error) error {
					if err != nil || !fi.Mode().IsDir() {
						return err
					}

					ok, err := filter(path, fi)

					if ok {
						z.Debugw("watching", "path", path)

						if err := watcher.Add(path); err != nil {
							return fmt.Errorf("watcher add: %w", err)
						}
					}

					return err
				}); err != nil {
					return fmt.Errorf("watcher walk: %w", err)
				}

				done := make(chan error)

				never := make(chan time.Time)

				go func() {
					var poll <-chan time.Time = never
					if indexOpts.interval != 0 {
						poll = time.After(indexOpts.interval)
					}

					if indexOpts.noINotify {
						<-poll
						done <- errRefresh
						return
					}

					for {
						select {
						case <-poll:
							done <- errRefresh
							return

						case event := <-watcher.Events:
							if event.Op&fsnotify.Write == fsnotify.Write {
								z.Infow("file modified", "event", event.Name)
								done <- errRefresh
								return
							}

						case err := <-watcher.Errors:
							done <- err
							return
						}
					}
				}()

				if err = <-done; err != nil && err != errRefresh {
					return fmt.Errorf("watcher: %w", err)
				}

				watcher.Close()
			}
		},
	}
)
