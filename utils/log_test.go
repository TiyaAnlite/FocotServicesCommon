package utils

import (
	"context"
	"fmt"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"os"
	"testing"
)

func TestLogging(t *testing.T) {
	ctx := context.Background()
	opelDSN := os.Getenv("TRACE_DSN")
	if opelDSN == "" {
		t.Error("openDSN not set")
		t.FailNow()
	}
	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(opelDSN),
		uptrace.WithServiceName("LoggingUtils"),
		uptrace.WithServiceVersion("test"),
		uptrace.WithDeploymentEnvironment("test"),
	)
	defer uptrace.Shutdown(ctx)
	tracer := otel.Tracer("LoggingUtils")
	ctx, span := tracer.Start(ctx, "TestLogging")
	defer span.End()
	testMsg := "testing: %s"
	InfoWithCtx(ctx, fmt.Sprintf(testMsg, "info"))
	WarningWithCtx(ctx, fmt.Sprintf(testMsg, "warning"))
	ErrorWithCtx(ctx, fmt.Sprintf(testMsg, "error"))
}
