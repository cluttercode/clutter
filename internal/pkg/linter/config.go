package linter

type Rule struct {
	Name       string   `json:"name"`
	PathGlob   string   `json:"path-glob"`
	PathRegexp string   `json:"path-re"`
	Shell      []string `json:"shell"`
	Expr       string   `json:"expr"`
}

type Config struct {
	Rules []Rule `json:"rules"`
}
