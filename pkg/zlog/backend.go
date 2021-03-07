package zlog

type Backend interface {
	Report(Level, string, []Pair)
	Named(string) Backend
	With([]Pair) Backend
}
