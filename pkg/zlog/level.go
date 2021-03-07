package zlog

import (
	"fmt"
	"strings"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
	Fatal
	Panic
	DebugPanic
	unknown
)

var (
	levelToString = map[Level]string{
		Debug:      "debug",
		Info:       "info",
		Warn:       "warn",
		Error:      "error",
		Fatal:      "fatal",
		Panic:      "panic",
		DebugPanic: "dpanic",
	}

	stringToLevel = map[string]Level{
		"debug":  Debug,
		"info":   Info,
		"warn":   Warn,
		"error":  Error,
		"fatal":  Fatal,
		"panic":  Panic,
		"dpanic": DebugPanic,
	}
)

func (l Level) String() string {
	if l.IsUnknown() {
		return fmt.Sprintf("%d", l)
	}

	return levelToString[l]
}

var ErrUnknownLevel = fmt.Errorf("unknown level")

func (l Level) IsUnknown() bool { return l >= unknown || l < Debug }

func ParseLevelString(txt string) (Level, error) {
	l, found := stringToLevel[strings.ToLower(txt)]
	if !found {
		return Debug, ErrUnknownLevel
	}

	return l, nil
}

func IsLevelReportable(min, asked Level) bool { return asked >= Fatal || asked >= min }
