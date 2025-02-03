package ai

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type aiLogger struct {
	*zap.Logger
}

func (l *aiLogger) Debugf(format string, v ...interface{}) {
	l.log(zap.DebugLevel, format, v...)
}

func (l *aiLogger) Errorf(format string, v ...interface{}) {
	l.log(zap.ErrorLevel, format, v...)
}

func (l *aiLogger) Infof(format string, v ...interface{}) {
	l.log(zap.InfoLevel, format, v...)
}

func (l *aiLogger) Warnf(format string, v ...interface{}) {
	l.log(zap.WarnLevel, format, v...)
}

func (l *aiLogger) log(Level zapcore.Level, format string, v ...interface{}) {
	if l.Level().Enabled(Level) {
		l.Check(Level, fmt.Sprintf(format, v...))
	}
}
