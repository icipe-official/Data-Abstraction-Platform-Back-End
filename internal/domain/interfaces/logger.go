package interfaces

import (
	"context"
	"log/slog"
)

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}
