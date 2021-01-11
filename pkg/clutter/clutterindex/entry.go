package clutterindex

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"github.com/cluttercode/clutter/internal/pkg/scanner"
)

type Attrs map[string]string

type AttrsStruct struct {
	as Attrs
}

func (a Attrs) ToStruct() *AttrsStruct {
	return &AttrsStruct{as: a}
}

func AttrToString(k, v string) string {
	if v == "" {
		return k
	}

	return fmt.Sprintf("%s=%s", k, v)
}

func (a AttrsStruct) Has(k string) bool {
	_, ok := a.as[k]
	return ok
}

func (a Attrs) Strings() []string {
	strs := make([]string, 0, len(a))

	for k, v := range a {
		strs = append(strs, AttrToString(k, v))
	}

	return strs
}

type Entry struct {
	Name  string
	Attrs Attrs
	Loc   scanner.Loc
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

	rs := []string{e.Name}

	attrs := make(map[string]string, len(e.Attrs)+1)
	for k, v := range e.Attrs {
		attrs[k] = v
	}

	ks := make([]string, 0, len(attrs)+1)

	for k := range e.Attrs {
		if k != "scope" { // this is handled separately below.
			ks = append(ks, k)
		}
	}

	sort.Strings(ks)

	// Loc is second only to scope.
	attrs["loc"] = e.Loc.String()

	ks = append([]string{"loc"}, ks...)

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

	e.Attrs = make(map[string]string, len(fs)-1)

	loc := false

	for _, f := range fs[1:] {
		parts := strings.SplitN(f, "=", 2)

		if parts[0] == "loc" {
			loc = true

			var locptr *scanner.Loc

			if locptr, err = scanner.ParseLocString(parts[1]); err != nil {
				return fmt.Errorf("loc: %w", err)
			}

			e.Loc = *locptr

			continue
		}

		if len(parts) == 1 {
			e.Attrs[parts[0]] = ""
		} else {
			e.Attrs[parts[0]] = parts[1]
		}
	}

	if !loc {
		return fmt.Errorf("no loc attr")
	}

	return nil
}
