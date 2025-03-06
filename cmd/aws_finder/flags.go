package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/pflag"
)

var _ pflag.Value = &logLevelFlag{}

type logLevelFlag struct {
	level slog.Level
}

func (l *logLevelFlag) String() string {
	return l.level.String()
}

func (l *logLevelFlag) Set(s string) error {
	return l.level.UnmarshalText([]byte(s))
}

func (l *logLevelFlag) Type() string {
	return fmt.Sprintf("%s|%s|%s|%s", slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError)
}
