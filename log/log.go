package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

type Config struct {

	// AsJSON serializes the core log entry data in a JSON format
	AsJSON bool

	// AsFile save log to file instead of print to stdout
	AsFile bool

	// UseRotator writing logs to rolling files
	UseRotator bool

	// Filename is the file to write logs to.
	Filename string

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool
}

func Configure(c Config) {
	if c.Filename == "" {
		c.Filename = "logs/grest.log"
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if c.AsJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	ws := zapcore.AddSync(os.Stdout)
	if c.AsFile {
		if c.UseRotator {
			ws = zapcore.AddSync(&lumberjack.Logger{
				Filename:   c.Filename,
				MaxSize:    c.MaxSize,
				MaxBackups: c.MaxBackups,
				MaxAge:     c.MaxAge,
				Compress:   c.Compress,
			})
		} else {
			f, _ := os.OpenFile(c.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			ws = zapcore.AddSync(f)
		}
	}

	core := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
	logger = zap.New(core, zap.AddCaller())
}

func New() *zap.Logger {
	return logger
}
