package utils

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"
	"runtime"
	"strings"
	"unsafe"
)

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
const (
	InfoLog    = "INFO"
	ErrorLog   = "ERROR"
	WarningLog = "WARNING"
)

func recordLog(ctx context.Context, prevDepth int, severity string, message string) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		funcName := "???"
		pc, file, line, ok := runtime.Caller(prevDepth + 2)
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

		eventAttributes := []attribute.KeyValue{
			attribute.String("log.severity", severity),
			attribute.String("log.message", message),
			attribute.String("code.function", funcName),
			attribute.String("code.filepath", file),
			attribute.Int("code.lineno", line),
		}
		if severity == ErrorLog {
			stack := make([]byte, 4<<10) // default stack length: 4kb
			length := runtime.Stack(stack, false)
			eventAttributes = append(eventAttributes, attribute.String("exception.stacktrace", unsafe.String(unsafe.SliceData(stack[:length]), length)))
		}
		span.AddEvent("log", trace.WithAttributes(eventAttributes...))
	}

	switch severity {
	case "INFO":
		klog.InfoDepth(prevDepth+2, message)
	case "ERROR":
		klog.ErrorDepth(prevDepth+2, message)
	case "WARNING":
		klog.WarningDepth(prevDepth+2, message)
	}
}

func InfoWithCtx(ctx context.Context, message string, prevDepth ...int) {
	depth := 0
	if len(prevDepth) > 0 {
		depth = prevDepth[0]
	}
	recordLog(ctx, depth, InfoLog, message)
}

func WarningWithCtx(ctx context.Context, message string, prevDepth ...int) {
	depth := 0
	if len(prevDepth) > 0 {
		depth = prevDepth[0]
	}
	recordLog(ctx, depth, WarningLog, message)
}

func ErrorWithCtx(ctx context.Context, message string, prevDepth ...int) {
	depth := 0
	if len(prevDepth) > 0 {
		depth = prevDepth[0]
	}
	recordLog(ctx, depth, ErrorLog, message)
}
