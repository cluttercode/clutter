package zlog

import (
	"fmt"
	"log"
	"os"
)

type GoLogBackend struct {
	Logger *log.Logger
	Level  Level
	pairs  Pairs
}

var _ Backend = &GoLogBackend{}

func NewDefaultBackend() *GoLogBackend {
	return &GoLogBackend{Logger: log.New(os.Stderr, "", 0)}
}

func (l *GoLogBackend) Named(name string) Backend {
	prefix := l.Logger.Prefix()

	if n := len(prefix); n == 0 {
		prefix = name
	} else if prefix[n-1] == ' ' {
		prefix = fmt.Sprintf("%s.%s", prefix[:n-2], name)
	} else {
		prefix = fmt.Sprintf("%s.%s", prefix, name)
	}

	prefix += " "

	return &GoLogBackend{
		Logger: log.New(l.Logger.Writer(), prefix, l.Logger.Flags()),
		Level:  l.Level,
		pairs:  l.pairs,
	}
}

func (l *GoLogBackend) With(pairs []Pair) Backend {
	return &GoLogBackend{
		Logger: l.Logger,
		Level:  l.Level,
		pairs:  append(l.pairs, pairs...),
	}
}

func (l *GoLogBackend) Report(lvl Level, msg string, pairs []Pair) {
	if !IsLevelReportable(l.Level, lvl) {
		return
	}

	out := DefaultReportOutput(lvl, msg, append(l.pairs, pairs...))

	switch lvl {
	case Fatal:
		l.Logger.Fatalln(out)
	case Panic:
		l.Logger.Panicln(out)
	case DebugPanic:
		if l.Level == Debug {
			l.Logger.Panicln(out)
		} else {
			l.Logger.Println(out)
		}
	default:
		l.Logger.Println(out)
	}
}
