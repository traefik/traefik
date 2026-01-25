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

func (l LogrusStdWrapper) Print(args ...any) {
	l.logger.Debug().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l LogrusStdWrapper) Printf(s string, args ...any) {
	l.logger.Debug().CallerSkipFrame(1).Msgf(s, args...)
}

func (l LogrusStdWrapper) Println(args ...any) {
	l.logger.Debug().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l LogrusStdWrapper) Fatal(args ...any) {
	l.logger.Fatal().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l LogrusStdWrapper) Fatalf(s string, args ...any) {
	l.logger.Fatal().CallerSkipFrame(1).Msgf(s, args...)
}

func (l LogrusStdWrapper) Fatalln(args ...any) {
	l.logger.Fatal().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l LogrusStdWrapper) Panic(args ...any) {
	l.logger.Panic().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l LogrusStdWrapper) Panicf(s string, args ...any) {
	l.logger.Panic().CallerSkipFrame(1).Msgf(s, args...)
}

func (l LogrusStdWrapper) Panicln(args ...any) {
	l.logger.Panic().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}
