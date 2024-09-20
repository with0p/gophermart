package logger

import "go.uber.org/zap"

var logger = zap.Must(zap.NewProduction()).Sugar()

func Info(text string) {
	logger.Infoln(
		"info", text,
	)
}

func Error(err error) {
	logger.Errorln(
		"ERROR", err.Error(),
	)
}
