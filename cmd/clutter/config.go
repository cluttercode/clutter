package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

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
	UseIndex bool           `yaml:"use-index"`
	Scanner  scanner.Config `yaml:"scanner"`
	Linter   linter.Config  `yaml:"linter"`
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

	if err := yaml.UnmarshalStrict(bs, &cfg); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	return nil
}
