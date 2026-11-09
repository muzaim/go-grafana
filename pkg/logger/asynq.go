package logger

import "go.uber.org/zap"

type AsynqLogger struct {
	logger *zap.SugaredLogger
}

func NewAsynqLogger(l *zap.Logger) *AsynqLogger {
	if l == nil {
		l = GetLogger()
	}
	return &AsynqLogger{logger: l.Sugar()}
}

func (a *AsynqLogger) Debug(args ...interface{}) {
	a.logger.Debug(args...)
}

func (a *AsynqLogger) Info(args ...interface{}) {
	a.logger.Info(args...)
}

func (a *AsynqLogger) Warn(args ...interface{}) {
	a.logger.Warn(args...)
}

func (a *AsynqLogger) Error(args ...interface{}) {
	a.logger.Error(args...)
}

func (a *AsynqLogger) Fatal(args ...interface{}) {
	a.logger.Fatal(args...)
}
