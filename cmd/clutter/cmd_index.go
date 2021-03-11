package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	cli "github.com/urfave/cli/v2"

	"github.com/cluttercode/clutter/internal/pkg/index"
	"github.com/cluttercode/clutter/internal/pkg/parser"
	"github.com/cluttercode/clutter/internal/pkg/scanner"
)

var (
	indexOpts = struct {
		watch, noINotify, print bool
		interval                time.Duration
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
			&cli.BoolFlag{
				Name:        "print",
				Aliases:     []string{"p"},
				Destination: &indexOpts.print,
			},
		},
		Aliases: []string{"i"},
		Usage:   "generate index database",
		Action: func(c *cli.Context) error {
			scan := func() error {
				z.Info("scanning")

				scan, err := scanner.NewScanner(z.Named("scanner"), cfg.Scanner)
				if err != nil {
					return fmt.Errorf("new scanner: %w", err)
				}

				elems, err := scan(".", func(e *scanner.RawElement) error {
					z.Infow("found", "element", e)
					return nil
				})

				if err != nil {
					return fmt.Errorf("scan: %w", err)
				}

				ents, err := parser.ParseElements(elems)
				if err != nil {
					return fmt.Errorf("parser: %w", err)
				}

				idx := index.NewIndex(ents)

				meta := fmt.Sprintf("%s %s", version, commit)

				if indexOpts.print {
					_ = index.WriteFile("stdout", idx, meta)
				}

				z.Infow("writing index", "n", idx.Size())

				if err := index.WriteFile(opts.indexPath, idx, meta); err != nil {
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

				watcher, err := fsnNewWatcher()
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
							if event.Op&(fsnWrite|fsnRemove|fsnRename|fsnCreate) != 0 {
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
