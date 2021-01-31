package clutterindex

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"github.com/cluttercode/clutter/internal/pkg/scanner"

	"github.com/cluttercode/clutter/pkg/strmatcher"
)

type Entry struct {
	Name  string
	Attrs Attrs
	Loc   scanner.Loc
}

func (e *Entry) IsSearch() (patternType string, yes bool) {
	patternType, yes = e.Attrs["search"]
	return
}

func (e *Entry) Matcher() (func(*Entry) bool, error) {
	pt, _ := e.IsSearch()

	var compile strmatcher.Compiler

	switch pt {
	case "exact":
		compile = strmatcher.CompileExactMatcher
	case "regexp":
		compile = strmatcher.CompileRegexpMatcher
	case "glob":
		compile = strmatcher.CompileGlobMatcher
	default:
		return nil, fmt.Errorf("unknown pattern type")
	}

	matchName := func(string) bool { return true }
	if e.Name != "" {
		var err error
		matchName, err = compile(e.Name)
		if err != nil {
			return nil, fmt.Errorf("name pattern error: %w", err)
		}
	}

	attrsMatchers := make(map[string]func(string) bool, len(e.Attrs))
	for k, v := range e.Attrs {
		if k == "search" {
			continue
		}

		m, err := compile(v)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q for %q:, %w", v, k, err)
		}

		attrsMatchers[k] = m
	}

	return func(other *Entry) bool {
		if _, search := other.IsSearch(); search {
			return false
		}

		if !matchName(other.Name) {
			return false
		}

		for k, m := range attrsMatchers {
			any := false

			for a, v := range other.AttrsWithLoc() {
				if a != k {
					continue
				}

				if any = m(v); any {
					break
				}
			}

			if !any {
				return false
			}
		}

		return true
	}, nil
}

func (e *Entry) IsReferredBy(them *Entry) bool {
	mine, theirs := e.Attrs["scope"], them.Attrs["scope"]

	if mine == "" {
		if theirs == "" {
			// mine == "", theirs == ""
			return true
		}

		// mine == "", theirs != ""
		if strings.HasSuffix(theirs, "/") {
			// is my path included in their scope?
			return strings.HasPrefix(e.Loc.Path, theirs)
		}

		// specific file
		return e.Loc.Path == theirs
	}

	// mine != ""

	if theirs == "" {
		theirs = them.Loc.Path
	}

	if strings.HasSuffix(mine, "/") {
		return strings.HasPrefix(theirs, mine)
	}

	// specific file
	return mine == theirs
}

func (e *Entry) String() string { return e.marshal() }

func (e *Entry) AttrsWithLoc() Attrs {
	attrs := make(Attrs, len(e.Attrs)+1)
	for k, v := range e.Attrs {
		attrs[k] = v
	}

	attrs["loc"] = e.Loc.String()

	return attrs
}

// Entries are marshalled in a way that a simple string sort of them will
// give the same result like we sort them in [# index-entry-sorting #].
func (e *Entry) marshal() string {
	b := &strings.Builder{}
	w := csv.NewWriter(b)
	w.Comma = ' '

	rs := []string{e.Name, e.Loc.String()}

	attrs := make(map[string]string, len(e.Attrs))
	ks := make([]string, 0, len(attrs))

	for k, v := range e.Attrs {
		attrs[k] = v

		if k != "scope" { // this is handled separately below.
			ks = append(ks, k)
		}
	}

	sort.Strings(ks)

	// Scope always first because it's important.
	if scope := e.Attrs["scope"]; scope != "" {
		ks = append([]string{"scope"}, ks...)
	}

	for _, k := range ks {
		rs = append(rs, AttrToString(k, attrs[k]))
	}

	_ = w.Write(rs)
	w.Flush()

	return strings.TrimSuffix(b.String(), "\n")
}

func (e *Entry) unmarshal(text string) error {
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = ' '

	fs, err := r.Read()
	if err != nil {
		return err
	}

	if len(fs) < 2 {
		return fmt.Errorf("invalid record")
	}

	e.Name = fs[0]

	var locptr *scanner.Loc

	if locptr, err = scanner.ParseLocString(fs[1]); err != nil {
		return fmt.Errorf("loc: %w", err)
	}

	e.Loc = *locptr

	e.Attrs = make(map[string]string, len(fs)-2)

	for _, f := range fs[2:] {
		parts := strings.SplitN(f, "=", 2)

		if len(parts) == 1 {
			e.Attrs[parts[0]] = ""
		} else {
			e.Attrs[parts[0]] = parts[1]
		}
	}

	return nil
}
