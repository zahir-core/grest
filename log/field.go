package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Bool(key string, val bool) zapcore.Field {
	return zap.Bool(key, val)
}

func ByteString(key string, val []byte) zapcore.Field {
	return zap.ByteString(key, val)
}

func Float64(key string, val float64) zapcore.Field {
	return zap.Float64(key, val)
}

func Int64(key string, val int64) zapcore.Field {
	return zap.Int64(key, val)
}

func String(key string, val string) zapcore.Field {
	return zap.String(key, val)
}

func Time(key string, val time.Time) zapcore.Field {
	return zap.Time(key, val)
}

func Duration(key string, val time.Duration) zapcore.Field {
	return zap.Duration(key, val)
}

func Any(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}
