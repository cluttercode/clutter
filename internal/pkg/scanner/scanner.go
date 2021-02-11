package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func NewScanner(z *zap.SugaredLogger, cfg Config) (func(root string, f func(*RawElement) error) ([]*RawElement, error), error) {
	filter, err := NewFilter(z, cfg)
	if err != nil {
		return nil, err
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

			stopped := false

			if err := Scan(z, cfg.Bracket, path, nil, func(elem *RawElement) error {
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
