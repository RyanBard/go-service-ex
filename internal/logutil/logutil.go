package logutil

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/RyanBard/go-service-ex/internal/ctxutil"
)

func LogAttrSVC(name string) slog.Attr {
	return slog.String("svc", name)
}

func LogAttrReqID(ctx context.Context) slog.Attr {
	reqID, _ := ctx.Value(ctxutil.ContextKeyReqID{}).(string)
	return slog.String("reqID", reqID)
}

func LogAttrFN(name string) slog.Attr {
	return slog.String("fn", name)
}

func LogAttrLoggedInUserID(ctx context.Context) slog.Attr {
	userID, _ := ctx.Value(ctxutil.ContextKeyUserID{}).(string)
	return slog.String("loggedInUserID", userID)
}

func LogAttrError(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func ParseLevel(level string) (slog.Leveler, error) {
	lvl := new(slog.LevelVar)
	switch strings.ToLower(level) {
	case "debug":
		lvl.Set(slog.LevelDebug)
	case "info":
		lvl.Set(slog.LevelInfo)
	case "warn", "warning":
		lvl.Set(slog.LevelWarn)
	case "error":
		lvl.Set(slog.LevelError)
	default:
		return nil, fmt.Errorf("unknown log level: %s", level)
	}
	return lvl, nil
}
