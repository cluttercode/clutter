package clutterindex

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cluttercode/clutter/pkg/strmatcher"
)

func NewEntriesFilter(c strmatcher.Compiler, names, attrs []string) (func(*Entry) bool, error) {
	nm, err := strmatcher.NewMatchers(c, names)
	if err != nil {
		return nil, fmt.Errorf("invalid name matchers: %w", err)
	}

	am, err := strmatcher.NewMatchers(c, attrs)
	if err != nil {
		return nil, fmt.Errorf("invalid name matchers: %w", err)
	}

	return func(ent *Entry) bool {
		if len(names) != 0 && !nm.Any(ent.Name) {
			return false
		}

		for _, a := range am {
			found := false

			for _, b := range ent.AttrsWithLoc().Strings() {
				if found = a(b); found {
					break
				}
			}

			if !found {
				return false
			}
		}

		return true
	}, nil
}

func SliceSource(index *Index) func() (*Entry, error) {
	rest := index.entries

	return func() (curr *Entry, _ error) {
		if len(rest) > 0 {
			curr, rest = rest[0], rest[1:]
		}
		return
	}
}

func FileSource(path string) (next func() (*Entry, error), done func(), err error) {
	var f *os.File

	if path == "stdin" || path == "-" {
		f = os.Stdin
	} else if f, err = os.Open(path); err != nil {
		return nil, nil, fmt.Errorf("open: %w", err)
	}

	done = func() { f.Close() }

	scanner := bufio.NewScanner(f)

	i := 1
	first := true

	next = func() (*Entry, error) {
		if !scanner.Scan() {
			return nil, nil
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			return next()
		}

		if first {
			if text != versionMarker {
				return nil, fmt.Errorf("missing or incompatible index version marker - please reindex")
			}

			first = false

			return next()
		}

		ent := &Entry{}
		if err := ent.unmarshal(text); err != nil {
			return nil, fmt.Errorf("index line %d: %w", i, err)
		}

		return ent, nil
	}

	return
}

func ForEach(next func() (*Entry, error), fn func(*Entry) error) error {
	_, err := Filter(next, func(ent *Entry) (bool, error) { return false, fn(ent) })
	return err
}

func Filter(next func() (*Entry, error), filter func(*Entry) (bool, error)) (*Index, error) {
	if filter == nil {
		filter = func(*Entry) (bool, error) { return true, nil }
	}

	var results []*Entry

	for {
		ent, err := next()
		if err != nil {
			return nil, fmt.Errorf("source: %w", err)
		}

		if ent == nil {
			break
		}

		incl, err := filter(ent)

		if err != nil {
			return nil, fmt.Errorf("filter: %w", err)
		}

		if incl {
			results = append(results, ent)
		}
	}

	return &Index{entries: results}, nil
}
