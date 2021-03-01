package scanner

import (
	"fmt"
	"regexp"
)

type BracketConfig struct {
	Left  string `yaml:"left"`
	Right string `yaml:"right"`
}

func (c *BracketConfig) OverrideWith(o BracketConfig) {
	if c.Left == "" && c.Right == "" {
		*c = o
	}
}

func (c *BracketConfig) Regexp() (*regexp.Regexp, error) {
	return regexp.Compile(
		fmt.Sprintf(
			`%s.+?%s`,
			regexp.QuoteMeta(c.Left),
			regexp.QuoteMeta(c.Right),
		),
	)
}

type Config struct {
	Bracket BracketConfig `yaml:"bracket"`
	Ignore  []string      `yaml:"ignore"`
}
