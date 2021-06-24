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

func Error(msg string, fields ...zap.Field){
	logger.Error(msg, fields...)
}

func Errorf(template string, args ...interface{}){
	logger.Sugar().Errorf(template, args...)
}

func Fatal(msg string, fields ...zap.Field){
	logger.Fatal(msg, fields...)
}

func Fatalf(template string, args ...interface{}){
	logger.Sugar().Fatalf(template, args...)
}
