package log

import (
	"io"
	"log/slog"
)

func New(w io.Writer) *slog.Logger {
	h := slog.NewTextHandler(w, nil)

	return slog.New(h)
}
