package logger

import (
	"os"
	"time"
	"web-service/base/conf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Caller() *zap.SugaredLogger {
	return zap.S().WithOptions(zap.AddCaller())
}

// customTimeEncoder 用于在日志中打印指定时区的时间
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	cst, _ := time.LoadLocation("Asia/Shanghai")
	enc.AppendString(t.In(cst).Format("2006-01-02 15:04:05"))
}

func InitLogger() {
	config := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime:    customTimeEncoder,
		EncodeCaller:  zapcore.FullCallerEncoder,
	}
	var logFormat = conf.GetLogFormat()
	var encoder zapcore.Encoder
	switch logFormat {
	case "json":
		encoder = zapcore.NewJSONEncoder(config)
	case "console":
		encoder = zapcore.NewConsoleEncoder(config)
	default:
		encoder = zapcore.NewConsoleEncoder(config)
	}

	writer := zapcore.AddSync(os.Stdout)
	var logLevelStr = conf.GetLogLevel()
	var logLevel zapcore.Level
	switch logLevelStr {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "err":
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}
	core := zapcore.NewCore(encoder, writer, logLevel)

	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
	zap.S().Infof("log initialization successful, log format: %s, log level: %s", logFormat, logLevelStr)
}
