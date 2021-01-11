package scanner

import (
	"fmt"
	"regexp"
	"strconv"
)

type Loc struct {
	Path   string
	Line   int
	Column int
}

type RawElement struct {
	Text string
	Loc  Loc
}

func (e *Loc) Less(other Loc) bool {
	if x, y := e.Path, other.Path; x != y {
		return e.Path < other.Path
	}

	if x, y := e.Line, other.Line; x != y {
		return x < y
	}

	return e.Column < other.Column

}

func (e Loc) String() string { return fmt.Sprintf("%s:%d.%d", e.Path, e.Line, e.Column) }

var locRegexp = regexp.MustCompile(`^(.+):([0-9]+)\.([0-9]+)$`)

func ParseLocString(text string) (*Loc, error) {
	ms := locRegexp.FindAllStringSubmatch(text, -1)
	if len(ms) != 1 || len(ms[0]) != 4 {
		return nil, fmt.Errorf("invalid")
	}

	var (
		loc = Loc{Path: ms[0][1]}
		err error
	)

	if loc.Line, err = strconv.Atoi(ms[0][2]); err != nil {
		return nil, fmt.Errorf("invalid line number")
	}

	if loc.Column, err = strconv.Atoi(ms[0][3]); err != nil {
		return nil, fmt.Errorf("invalid column")
	}

	return &loc, nil
}
