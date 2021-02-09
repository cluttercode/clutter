package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

var (
	opts = struct {
		logLevel   string
		verbose    bool
		debug      bool
		indexPath  string
		configPath string
		nocolor    bool
	}{
		logLevel:   "info",
		indexPath:  configPath(indexFilename),
		configPath: configPath(configFilename),
	}

	indexFlag = cli.StringFlag{
		Name:        "index-path",
		Aliases:     []string{"i"},
		Value:       opts.indexPath,
		Destination: &opts.indexPath,
	}

	app = cli.App{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "nocolor",
				Aliases:     []string{"nc"},
				Destination: &opts.nocolor,
				Usage:       "do not colorize logs",
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &opts.verbose,
				Usage:       "sets log level to info. overrides log-level.",
			},
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d", "vv"},
				Destination: &opts.debug,
				Usage:       "sets log level to debug. overrides log-level and verbose.",
			},
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"ll"},
				Value:       "warn",
				Destination: &opts.logLevel,
			},
			&cli.StringFlag{
				Name:        "config-path",
				Aliases:     []string{"c"},
				Value:       opts.configPath,
				Destination: &opts.configPath,
			},
			&indexFlag,
		},
		Commands: []*cli.Command{
			&indexCommand,
			&lintCommand,
			&searchCommand,
			&resolveCommand,
			&versionCommand,
		},
		Before: func(c *cli.Context) error {
			if err := loadConfig(opts.configPath); err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			level := opts.logLevel

			if opts.verbose {
				level = "info"
			}

			if opts.debug {
				level = "debug"
			}

			if err := initLogger(level, !opts.nocolor); err != nil {
				return fmt.Errorf("init logger: %w", err)
			}

			z.Debugw("started", "cfg", cfg)

			return nil
		},
	}
)

func indexPaths(c *cli.Context) []string {
	if c.IsSet(indexFlag.Name) {
		return []string{opts.indexPath}
	}

	if cfg.IgnoreIndex || opts.indexPath == "" {
		return []string{""}
	}

	return []string{opts.indexPath, ""}
}
