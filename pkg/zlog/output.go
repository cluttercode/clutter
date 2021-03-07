package zlog

import (
	"fmt"
)

func DefaultReportOutput(lvl Level, msg string, pairs []Pair) string {
	out := fmt.Sprintf("[%v] %s", lvl, msg)

	if len(pairs) > 0 {
		out = fmt.Sprintf("%s {%v}", out, Pairs(pairs))
	}

	return out
}
