package zlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func New(runMode string) {
	cfg := zap.NewProductionConfig()
	if runMode == "development" {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	defer Sync()

	var err error
	L, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

func Sync() error {
	return L.Sync()
}

// Helper functions for structured logging

// String returns a string field for structured logging
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int returns an int field for structured logging
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int64 returns an int64 field for structured logging
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Uint64 returns a uint64 field for structured logging
func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

// Error returns an error field for structured logging
func Error(err error) zap.Field {
	return zap.Error(err)
}

// Bool returns a boolean field for structured logging
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Any returns a field that encodes anything
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}
