package tracex

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"testing"
	"time"
)

func TestServiceTraceHelper(t *testing.T) {
	ctx := context.Background()
	cfg := &ServiceTraceHelper{}
	cfg.SetupTrace()
	defer cfg.Shutdown(ctx)
	tracer := cfg.NewTracer()
	wg := sync.WaitGroup{}

	ctx, main := tracer.Start(ctx, "UnitTest")
	defer main.End()

	go func() {
		_, spanGo1 := tracer.Start(ctx, "UniTest-Go-1")
		defer spanGo1.End()
		wg.Add(1)
		defer wg.Done()
		time.Sleep(time.Second)
	}()

	go func() {
		_, spanGo2 := tracer.Start(ctx, "UniTest-Go-2")
		defer spanGo2.End()
		wg.Add(1)
		defer wg.Done()
		time.Sleep(time.Second * 3)
	}()

	_, span1 := tracer.Start(ctx, "UnitTest-1")
	span1.SetAttributes(attribute.String("unitTest.step", "1"))
	time.Sleep(time.Millisecond * 300)
	span1.End()

	_, span2 := tracer.Start(ctx, "UnitTest-2")
	span2.SetAttributes(attribute.String("unitTest.step", "2"))
	span2.AddEvent("small.step")
	time.Sleep(time.Millisecond * 100)
	span2.AddEvent("small.step")
	time.Sleep(time.Millisecond * 300)
	span2.End()

	_, span3 := tracer.Start(ctx, "UnitTest-waiter")
	span3.SetAttributes(attribute.String("unitTest.step", "waiting"))
	wg.Wait()
	span3.End()
}

func TestRemoteTraceHelper(t *testing.T) {
	ctx := context.Background()
	cfg := &ServiceTraceHelper{}
	cfg.SetupTrace()
	defer cfg.Shutdown(ctx)
	tracer := cfg.NewTracer()

	ctx, main := tracer.Start(ctx, "UnitTest")
	defer main.End()
	_, localSpan := tracer.Start(ctx, "LocalSpan")
	defer localSpan.End()
	t.Log("LocalSpan start")
	t.Logf("traceID: %s", localSpan.SpanContext().TraceID())
	t.Logf("spanID: %s", localSpan.SpanContext().SpanID())
	t.Logf("traceFlags: %s", localSpan.SpanContext().TraceFlags())
	t.Logf("traceState: %s", localSpan.SpanContext().TraceState())
	time.Sleep(time.Millisecond * 300)
	localContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    localSpan.SpanContext().TraceID(),
		SpanID:     localSpan.SpanContext().SpanID(),
		TraceFlags: localSpan.SpanContext().TraceFlags(),
		TraceState: localSpan.SpanContext().TraceState(),
		Remote:     true,
	})
	_, remoteSpan := tracer.Start(trace.ContextWithSpanContext(context.Background(), localContext), "RemoteSpan")
	defer remoteSpan.End()
	t.Log("RemoteSpan start")
	time.Sleep(time.Millisecond * 500)
	t.Log("RemoteSpan end")
}
