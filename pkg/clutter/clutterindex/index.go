package clutterindex

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const versionMarker = "# v4"

type Index struct{ entries []*Entry }

// [# index-entry-sorting #] sorts by name, then loc.

type sorter []*Entry

func (i sorter) Len() int      { return len(i) }
func (i sorter) Swap(a, b int) { i[a], i[b] = i[b], i[a] }
func (i sorter) Less(a, b int) bool {
	aa, bb := i[a], i[b]

	if aa.Name == bb.Name {
		return aa.Loc.Less(bb.Loc)
	}

	return aa.Name < bb.Name
}

func NewIndex(ents []*Entry) *Index { return (&Index{}).Add(ents) }

// Add modifies i.
func (i *Index) Add(ents []*Entry) *Index {
	i.entries = append(i.entries, ents...)
	sort.Sort(sorter(i.entries))
	return i
}

func (i *Index) Size() int { return len(i.entries) }

func Write(path string, index *Index, comment string) error {
	var (
		f      *os.File
		commit = func() error { return nil }
	)

	if path == "stdout" || path == "-" {
		f = os.Stdout
	} else {
		dir := filepath.Dir(path)
		if dir != "." && dir != ".." && dir != "/" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("mkdir config path: %w", err)
			}
		}

		tmpPath := path + ".next"

		defer os.Remove(tmpPath)

		var err error

		if f, err = os.Create(tmpPath); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		commit = func() error {
			if err := os.Rename(tmpPath, path); err != nil {
				return fmt.Errorf("move: %w", err)
			}

			return nil
		}
	}

	fmt.Fprintf(f, "%s %s\n", versionMarker, comment)

	for _, i := range index.entries {
		text := i.marshal() + "\n"

		if _, err := f.WriteString(text); err != nil {
			f.Close()
			return fmt.Errorf("write: %w", err)
		}
	}

	f.Close()

	return commit()
}
