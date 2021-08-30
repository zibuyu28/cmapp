package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var logger *zap.Logger

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l
}

func InitCus() {
	consoleConfig := zapcore.EncoderConfig{
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	// // 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	skip := zap.AddCallerSkip(2)
	// // trace
	//trace := zap.AddStacktrace(zap.ErrorLevel)
	// 开启文件及行号
	development := zap.Development()

	atomicLevel := zap.NewAtomicLevel()
	var l zapcore.Level
	_ = l.UnmarshalText([]byte("debug"))
	atomicLevel.SetLevel(l)

	cc := zapcore.NewCore(zapcore.NewConsoleEncoder(consoleConfig), zapcore.AddSync(os.Stdout), atomicLevel)
	logger = zap.New(cc, caller,skip, development)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Debugf(ctx context.Context, template string, args ...interface{}) {
	logger.Sugar().Debugf(template, args...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Infof(ctx context.Context, template string, args ...interface{}) {
	logger.Sugar().Infof(template, args...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Errorf(ctx context.Context, template string, args ...interface{}) {
	logger.Sugar().Errorf(template, args...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Warnf(ctx context.Context, template string, args ...interface{}) {
	logger.Sugar().Warnf(template, args...)
}


func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Fatalf(ctx context.Context, template string, args ...interface{}) {
	logger.Sugar().Fatalf(template, args...)
}
