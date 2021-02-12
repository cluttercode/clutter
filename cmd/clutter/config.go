package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/cluttercode/clutter/internal/pkg/linter"
	"github.com/cluttercode/clutter/internal/pkg/scanner"
)

const (
	defaultClutterDir = ".clutter"
	configFilename    = "config.yaml"
	indexFilename     = "index"
)

func configPath(p string) string { return filepath.Join(defaultClutterDir, p) }

type config struct {
	UseIndex bool           `json:"use-index"`
	Scanner  scanner.Config `json:"scanner"`
	Linter   linter.Config  `json:"linter"`
}

var (
	defaultCfg = config{
		Scanner: scanner.Config{
			Bracket: scanner.BracketConfig{
				Left:  "[#",
				Right: "#]",
			},
		},
	}

	cfg = defaultCfg
)

func loadConfig(path string) error {
	bs, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err := yaml.UnmarshalStrict(bs, &cfg, yaml.DisallowUnknownFields); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	return nil
}
