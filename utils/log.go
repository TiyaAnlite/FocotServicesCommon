package utils

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"
	"runtime"
	"strings"
)

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
const (
	InfoLog    = "INFO"
	ErrorLog   = "ERROR"
	WarningLog = "WARNING"
)

func recordLog(ctx context.Context, severity string, message string) {
	funcName := "???"
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 1
	}
	if slash := strings.LastIndex(file, "/"); slash >= 0 {
		file = file[slash+1:]
	}
	if f := runtime.FuncForPC(pc); f != nil {
		funcName = f.Name()
		if slash := strings.LastIndex(funcName, "/"); slash >= 0 {
			funcName = funcName[slash+1:]
		}
	}
	span := trace.SpanFromContext(ctx)
	span.AddEvent("log", trace.WithAttributes(
		attribute.String("log.severity", severity),
		attribute.String("log.message", message),
		attribute.String("code.function", funcName),
		attribute.String("code.filepath", file),
		attribute.Int("code.lineno", line),
	))
	span.End()
	switch severity {
	case "INFO":
		klog.InfoDepth(2, message)
	case "ERROR":
		klog.ErrorDepth(2, message)
	case "WARNING":
		klog.WarningDepth(2, message)
	}
}

func InfoWithCtx(ctx context.Context, message string) {
	recordLog(ctx, InfoLog, message)
}

func WarningWithCtx(ctx context.Context, message string) {
	recordLog(ctx, WarningLog, message)
}

func ErrorWithCtx(ctx context.Context, message string) {
	recordLog(ctx, ErrorLog, message)
}
