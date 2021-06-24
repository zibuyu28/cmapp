package log

import "go.uber.org/zap"

var logger *zap.Logger

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Debugf(template string, args ...interface{}){
	logger.Sugar().Debugf(template, args...)
}

func Info(msg string, fields ...zap.Field){
	logger.Info(msg, fields...)
}

func Infof(template string, args ...interface{}){
	logger.Sugar().Infof(template, args...)
}