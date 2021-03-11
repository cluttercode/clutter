// +build nofsnotify

package main

import "errors"

type Op uint32

const (
	fsnCreate Op = iota
	fsnWrite
	fsnRemove
	fsnRename
)

type Event struct {
	Name string
	Op   Op
}

type Watcher struct {
	Errors chan error
	Events chan Event
}

func (w *Watcher) Add(name string) error { return nil }
func (w *Watcher) Close() error          { return nil }

func fsnNewWatcher() (*Watcher, error) {
	return nil, errors.New("fsnotify is not supported")
}
