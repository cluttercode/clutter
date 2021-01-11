package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var z *zap.SugaredLogger = zap.NewNop().Sugar()

func initLogger(level string) error {
	zcfg := zap.NewDevelopmentConfig()

	zcfg.DisableStacktrace = true
	zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zcfg.EncoderConfig.EncodeTime = func(time.Time, zapcore.PrimitiveArrayEncoder) {}

	if level != "debug" {
		zcfg.EncoderConfig.EncodeDuration = nil
		zcfg.EncoderConfig.EncodeCaller = nil
	}

	if err := zcfg.Level.UnmarshalText([]byte(level)); err != nil {
		return fmt.Errorf(`invalid log level "%s": %w`, level, err)
	}

	zz, err := zcfg.Build(zap.AddCaller())
	if err != nil {
		return fmt.Errorf("failed initializing log: %w", err)
	}

	z = zz.Sugar()

	return nil
}
