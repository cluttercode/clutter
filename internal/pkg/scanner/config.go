package scanner

import (
	"fmt"
	"regexp"
)

type BracketConfig struct {
	Left  string `json:"left"`
	Right string `json:"right"`
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

type ToolConfig struct {
	Pattern string                 `json:"pattern"`
	Tool    string                 `json:"tool"`
	Params  map[string]interface{} `json:"params"`
	Bracket BracketConfig          `json:"bracket"`
}

type Config struct {
	Bracket BracketConfig `json:"bracket"`
	Ignore  []string      `json:"ignore"`
	Tools   []ToolConfig  `json:"tools"`
}
