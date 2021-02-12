package index

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadFile(path string) (*Index, error) {
	var (
		f   *os.File
		err error
	)

	if path == "stdin" || path == "-" {
		f = os.Stdin
	} else if f, err = os.Open(path); err != nil {
		return nil, err // don't wrap here - checking for IsNotExist in caller.
	}

	scanner := bufio.NewScanner(f)

	first := true
	ents := make([]*Entry, 0, 10)

	for i := 1; scanner.Scan(); i++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		if first {
			if !strings.HasPrefix(text, versionMarker+" ") {
				return nil, fmt.Errorf("missing or incompatible index version marker - please reindex")
			}

			first = false

			continue
		}

		ent := &Entry{}
		if err := ent.unmarshal(text); err != nil {
			return nil, fmt.Errorf("index line %d: %w", i, err)
		}

		ents = append(ents, ent)
	}

	return NewIndex(ents), nil
}

var ErrStop = fmt.Errorf("stop")

func ForEach(idx *Index, fn func(*Entry) error) error {
	_, err := Filter(idx, func(ent *Entry) (bool, error) { return false, fn(ent) })
	return err
}

func Filter(idx *Index, filter func(*Entry) (bool, error)) (*Index, error) {
	if filter == nil {
		return idx, nil
	}

	var results []*Entry

	for _, ent := range idx.entries {
		incl, err := filter(ent)

		if err == ErrStop {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("filter: %w", err)
		}

		if incl {
			results = append(results, ent)
		}
	}

	return &Index{entries: results}, nil
}
