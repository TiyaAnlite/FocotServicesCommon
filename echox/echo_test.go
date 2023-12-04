package echox

import (
	"context"
	"github.com/TiyaAnlite/FocotServicesCommon/utils"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
	"os"
	"testing"
	"time"
)

func TestOpel(t *testing.T) {
	ctx := context.Background()
	opelDSN := os.Getenv("TRACE_DSN")
	if opelDSN == "" {
		t.Error("openDSN not set")
		t.FailNow()
	}
	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(opelDSN),
		uptrace.WithServiceName("EchoHelper"),
		uptrace.WithServiceVersion("test"),
		uptrace.WithDeploymentEnvironment("test"),
	)
	defer uptrace.Shutdown(ctx)
	traceFunc := func(c echo.Context) error {
		var ch trace.Span
		_, ch = RootTracer(c, "processing")
		t.Logf("do something...")
		time.Sleep(time.Second)
		ch.End()
		_, ch = RootTracer(c, "processing2")
		time.Sleep(time.Millisecond * 500)
		ch.End()
		t.Logf("ok")
		return NormalEmptyResponse(c)
	}
	go Run(&EchoConfig{
		Port: 8080,
	}, func(e *echo.Echo) {
		e.Use(otelecho.Middleware("EchoHelperTest"))
		e.Any("/", traceFunc)
	})
	utils.Wait4CtrlC()
}
