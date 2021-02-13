package scanner

import (
	"fmt"
	"regexp"
	"strconv"
)

type Loc struct {
	Path        string
	Line        int
	StartColumn int
	EndColumn   int
}

func (l *Loc) Less(other Loc) bool {
	if x, y := l.Path, other.Path; x != y {
		return l.Path < other.Path
	}

	if x, y := l.Line, other.Line; x != y {
		return x < y
	}

	return l.StartColumn < other.StartColumn

}

func (l Loc) Contains(other Loc) bool {
	return l.Path == other.Path &&
		l.Line == other.Line &&
		other.StartColumn >= l.StartColumn &&
		other.EndColumn <= l.EndColumn
}

func (l Loc) String() string {
	return fmt.Sprintf("%s:%d.%d-%d", l.Path, l.Line, l.StartColumn, l.EndColumn)
}

var locRegexp = regexp.MustCompile(`^(.+):([0-9]+)\.([0-9]+)(-[0-9]+)?$`)

func ParseLocString(text string) (*Loc, error) {
	ms := locRegexp.FindAllStringSubmatch(text, -1)
	if len(ms) != 1 || len(ms[0]) < 5 {
		return nil, fmt.Errorf("invalid")
	}

	var (
		loc = Loc{Path: ms[0][1]}
		err error
	)

	if loc.Line, err = strconv.Atoi(ms[0][2]); err != nil {
		return nil, fmt.Errorf("invalid line number")
	}

	if loc.StartColumn, err = strconv.Atoi(ms[0][3]); err != nil {
		return nil, fmt.Errorf("invalid start column")
	}

	if rest := ms[0][4]; rest != "" {
		if loc.EndColumn, err = strconv.Atoi(rest[1:]); err != nil {
			return nil, fmt.Errorf("invalid end column")
		}

		if loc.EndColumn <= loc.StartColumn {
			return nil, fmt.Errorf("invalid end column")
		}
	} else {
		loc.EndColumn = loc.StartColumn + 1
	}

	return &loc, nil
}
