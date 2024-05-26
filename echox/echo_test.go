package echox

import (
	"context"
	"github.com/TiyaAnlite/FocotServicesCommon/envx"
	"github.com/TiyaAnlite/FocotServicesCommon/utils"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
	"os"
	"os/signal"
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

func TestEchoServer(t *testing.T) {
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
	cfg := &EchoConfig{
		Port:              8080,
		UseHealthCheck:    false,
		TelemetryHostName: "EchoHelperTest",
	}
	envx.MustLoadEnv(cfg)
	go Run(cfg, func(e *echo.Echo) {
		e.Any("/", func(c echo.Context) error {
			return NormalResponse(c, "ok")
		})
	})
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	go func() {
		time.Sleep(time.Second * 5)
		ctrlc <- os.Interrupt
	}()
	<-ctrlc
	Shutdown(cfg)
}
