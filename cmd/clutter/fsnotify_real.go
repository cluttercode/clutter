// +build !nofsnotify

package main

import (
	"github.com/fsnotify/fsnotify"
)

var (
	fsnCreate = fsnotify.Create
	fsnWrite  = fsnotify.Write
	fsnRemove = fsnotify.Remove
	fsnRename = fsnotify.Rename
)

var fsnNewWatcher = fsnotify.NewWatcher
