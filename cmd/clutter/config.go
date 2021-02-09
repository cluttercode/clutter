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
	IgnoreIndex bool           `json:"ignore-index"`
	Scanner     scanner.Config `json:"scanner"`
	Linter      linter.Config  `json:"linter"`
}

var (
	defaultCfg = config{
		Scanner: scanner.Config{
			Bracket: scanner.BracketConfig{
				Left:  "[#",
				Right: "#]",
			},
			Tools: []scanner.ToolConfig{
				{
					Pattern: "*",
					Tool:    "builtin",
				},
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

	if err := yaml.Unmarshal(bs, &cfg); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	return nil
}
