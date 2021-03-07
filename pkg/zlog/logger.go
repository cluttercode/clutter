package zlog

import (
	"fmt"
)

type Logger struct{ Backend }

func NewDefaultLogger() *Logger { return &Logger{Backend: NewDefaultBackend()} }

func (l *Logger) invalid(op string, err error) {
	l.Backend.Report(DebugPanic, "invalid log usage", []Pair{{K: "op", V: op}, {K: "err", V: err}})
}

func (l *Logger) With(kvs ...interface{}) *Logger {
	pairs, err := toPairs(kvs)
	if err != nil {
		l.invalid("with", err)
		return l
	}

	return &Logger{Backend: l.Backend.With(pairs)}
}

func (l *Logger) Named(name string) *Logger { return &Logger{Backend: l.Backend.Named(name)} }

func (l *Logger) Report(lvl Level, msg string, pairs []Pair) {
	l.Backend.Report(lvl, msg, pairs)
}

func (l *Logger) report(lvl Level, msg string, kvs []interface{}) {
	pairs, err := toPairs(kvs)
	if err != nil {
		l.invalid("report", err)
		return
	}

	l.Backend.Report(lvl, msg, pairs)
}

// Each report must be done witht the same number of calls on the stack.
// This allows for a backend to just ignore the top two levels to find the call site.

const CallstackSkipCount = 2

func (l *Logger) debug(msg string, kvs []interface{})    { l.report(Debug, msg, kvs) }
func (l *Logger) Debugw(msg string, kvs ...interface{})  { l.debug(msg, kvs) }
func (l *Logger) Debug(msg string)                       { l.debug(msg, nil) }
func (l *Logger) Debugf(msg string, args ...interface{}) { l.debug(fmt.Sprintf(msg, args...), nil) }

func (l *Logger) info(msg string, kvs []interface{})    { l.report(Info, msg, kvs) }
func (l *Logger) Infow(msg string, kvs ...interface{})  { l.info(msg, kvs) }
func (l *Logger) Info(msg string)                       { l.info(msg, nil) }
func (l *Logger) Infof(msg string, args ...interface{}) { l.info(fmt.Sprintf(msg, args...), nil) }

func (l *Logger) warn(msg string, kvs []interface{})    { l.report(Warn, msg, kvs) }
func (l *Logger) Warnw(msg string, kvs ...interface{})  { l.warn(msg, kvs) }
func (l *Logger) Warn(msg string)                       { l.warn(msg, nil) }
func (l *Logger) Warnf(msg string, args ...interface{}) { l.warn(fmt.Sprintf(msg, args...), nil) }

func (l *Logger) err(msg string, kvs []interface{})      { l.report(Error, msg, kvs) }
func (l *Logger) Errorw(msg string, kvs ...interface{})  { l.err(msg, kvs) }
func (l *Logger) Error(msg string)                       { l.err(msg, nil) }
func (l *Logger) Errorf(msg string, args ...interface{}) { l.err(fmt.Sprintf(msg, args...), nil) }

func (l *Logger) fatal(msg string, kvs []interface{})    { l.report(Fatal, msg, kvs) }
func (l *Logger) Fatalw(msg string, kvs ...interface{})  { l.fatal(msg, kvs) }
func (l *Logger) Fatal(msg string)                       { l.fatal(msg, nil) }
func (l *Logger) Fatalf(msg string, args ...interface{}) { l.fatal(fmt.Sprintf(msg, args...), nil) }

func (l *Logger) panpanpan(msg string, kvs []interface{}) { l.report(Panic, msg, kvs) }
func (l *Logger) Panicw(msg string, kvs ...interface{})   { l.panpanpan(msg, kvs) }
func (l *Logger) Panic(msg string)                        { l.panpanpan(msg, nil) }
func (l *Logger) Panicf(msg string, args ...interface{})  { l.panpanpan(fmt.Sprintf(msg, args...), nil) }

func (l *Logger) dpanic(msg string, kvs []interface{})       { l.report(DebugPanic, msg, kvs) }
func (l *Logger) DebugPanicw(msg string, kvs ...interface{}) { l.dpanic(msg, kvs) }
func (l *Logger) DebugPanic(msg string)                      { l.dpanic(msg, nil) }
func (l *Logger) DebugPanicf(msg string, args ...interface{}) {
	l.dpanic(fmt.Sprintf(msg, args...), nil)
}
