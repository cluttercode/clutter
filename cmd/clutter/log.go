package main

import (
	"github.com/cluttercode/clutter/pkg/zlog"
)

var z *zlog.Logger = zlog.NewNopLogger()

func initLogger(level string, color bool) error {
	b := zlog.NewDefaultBackend()

	lvl, err := zlog.ParseLevelString(level)
	if err != nil {
		return err
	}

	b.Level = lvl

	z = &zlog.Logger{Backend: b}

	return nil
}
