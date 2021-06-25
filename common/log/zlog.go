package log

import (
	"context"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l
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
