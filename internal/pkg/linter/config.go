package linter

type Rule struct {
	Name       string   `yaml:"name"`
	PathGlob   string   `yaml:"path-glob"`
	PathRegexp string   `yaml:"path-re"`
	Shell      []string `yaml:"shell"`
}

type Config struct {
	Rules []Rule `yaml:"rules"`
}
