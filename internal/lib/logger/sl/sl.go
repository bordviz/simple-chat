package sl

import (
	"fmt"
	"log/slog"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func OpErr(op string, err error) slog.Attr {
	opErr := fmt.Errorf("%s: %w", op, err)
	return Err(opErr)
}
