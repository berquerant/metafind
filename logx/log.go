package logx

import (
	"encoding/json"
	"io"
	"log/slog"
)

func Setup(w io.Writer, level slog.Leveler) {
	setupLogger(w, level)
}

func setupLogger(w io.Writer, level slog.Leveler) {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func Err(err error) slog.Attr {
	return slog.Any("err", err)
}

func Jsonify(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func JSON(k string, v any) slog.Attr {
	return slog.String(k, string(Jsonify(v)))
}

func Panic(err error) {
	panic(err)
}

func PanicOnError(err error) {
	if err != nil {
		Panic(err)
	}
}
