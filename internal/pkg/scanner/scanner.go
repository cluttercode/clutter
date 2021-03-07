package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cluttercode/clutter/pkg/zlog"
)

func NewScanner(z *zlog.Logger, cfg Config) (func(root string, f func(*RawElement) error) ([]*RawElement, error), error) {
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

			if err := ScanFile(z, cfg.Bracket, path, func(elem *RawElement) error {
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
