package logs

import (
	"github.com/rs/zerolog"
)

type LogrusStdWrapper struct {
	logger zerolog.Logger
}

func NewLogrusWrapper(logger zerolog.Logger) *LogrusStdWrapper {
	return &LogrusStdWrapper{logger: logger}
}

func (l LogrusStdWrapper) Print(args ...interface{}) {
	l.logger.Print(args...)
}

func (l LogrusStdWrapper) Printf(s string, args ...interface{}) {
	l.logger.Printf(s, args...)
}

func (l LogrusStdWrapper) Println(args ...interface{}) {
	l.logger.Print(args...)
}

func (l LogrusStdWrapper) Fatal(args ...interface{}) {
	l.logger.Fatal().MsgFunc(MsgFunc(args...))
}

func (l LogrusStdWrapper) Fatalf(s string, args ...interface{}) {
	l.logger.Fatal().Msgf(s, args...)
}

func (l LogrusStdWrapper) Fatalln(args ...interface{}) {
	l.logger.Fatal().MsgFunc(MsgFunc(args...))
}

func (l LogrusStdWrapper) Panic(args ...interface{}) {
	l.logger.Panic().MsgFunc(MsgFunc(args...))
}

func (l LogrusStdWrapper) Panicf(s string, args ...interface{}) {
	l.logger.Panic().Msgf(s, args...)
}

func (l LogrusStdWrapper) Panicln(args ...interface{}) {
	l.logger.Panic().MsgFunc(MsgFunc(args...))
}
