package config

import "log/slog"

var LogLevels = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
}

var LogLevelsMap = map[string]slog.Level{
	LogLevels[0]: slog.LevelDebug,
	LogLevels[1]: slog.LevelInfo,
	LogLevels[2]: slog.LevelWarn,
	LogLevels[3]: slog.LevelError,
}
