package utils

import (
	"context"
	"fmt"
	"testing"
)

func TestLogging(t *testing.T) {
	ctx := context.Background()
	testMsg := "testing: %s"
	InfoWithCtx(ctx, fmt.Sprintf(testMsg, "info"))
	WarningWithCtx(ctx, fmt.Sprintf(testMsg, "warning"))
	ErrorWithCtx(ctx, fmt.Sprintf(testMsg, "error"))
}
