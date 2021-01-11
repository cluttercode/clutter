package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"go.uber.org/zap"
)

func NewScanner(z *zap.SugaredLogger, cfg Config) (func(root string, f func(*RawElement) error) ([]*RawElement, error), error) {
	filter, err := NewFilter(z, cfg)
	if err != nil {
		return nil, err
	}

	tools := make([]func(string) toolFunc, len(cfg.Tools))

	for i, tc := range cfg.Tools {
		if err := func(tc ToolConfig) error {
			glob, err := glob.Compile(tc.Pattern)
			if err != nil {
				return fmt.Errorf("invalid type glob: \"%s\" -> %w", tc.Pattern, err)
			}

			tc.Bracket.OverrideWith(cfg.Bracket)

			tool, err := makeTool(tc)
			if err != nil {
				return fmt.Errorf("error making tool \"%s\": %w", tc.Tool, err)
			}

			tools[i] = func(path string) toolFunc {
				if glob.Match(path) {
					return tool
				}

				return nil
			}

			return nil
		}(tc); err != nil {
			return nil, err
		}
	}

	return func(root string, f func(*RawElement) error) ([]*RawElement, error) {
		if f == nil {
			f = func(*RawElement) error { return nil }
		}

		var elems []*RawElement

		if err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			z := z.With("path", path)

			if err != nil {
				return fmt.Errorf("path %q: %w", path, err)
			}

			if include, err := filter(path, fi); !include {
				return err
			}

			z.Debug("reading")

			var tool toolFunc

			for _, tm := range tools {
				if tool = tm(path); tool != nil {
					break
				}
			}

			if tool == nil {
				z.Debug("no tool configured for file")
				return nil
			}

			stopped := false

			if err := tool(path, func(elem *RawElement) error {
				if strings.HasPrefix(elem.Text, "%") {
					switch elem.Text[1:] {
					case "stop":
						if stopped {
							return fmt.Errorf("already stopped")
						}

						stopped = true
					case "cont":
						if !stopped {
							return fmt.Errorf("not stopped")
						}

						stopped = false
					default:
						return fmt.Errorf("unknown pragma: %s", elem.Text)
					}

					return nil
				}

				if stopped {
					return nil
				}

				if err := f(elem); err != nil {
					return err
				}

				elems = append(elems, elem)

				return nil
			}); err != nil {
				return fmt.Errorf("file %s: tool: %w", path, err)
			}

			return nil
		}); err != nil {
			return nil, err
		}

		return elems, nil
	}, nil
}
