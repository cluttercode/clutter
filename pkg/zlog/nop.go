package zlog

type nopBackend struct{}

func NewNopBackend() Backend { return &nopBackend{} }

func NewNopLogger() *Logger { return &Logger{Backend: NewNopBackend()} }

func (n *nopBackend) Named(string) Backend { return n }

func (n *nopBackend) With([]Pair) Backend { return n }

func (*nopBackend) Report(lvl Level, msg string, pairs []Pair) {
	if lvl == Panic {
		panic(DefaultReportOutput(lvl, msg, pairs))
	}
}
