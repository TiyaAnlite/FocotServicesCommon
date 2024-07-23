package dbx

import (
	"context"
	"github.com/TiyaAnlite/FocotServicesCommon/envx"
	"github.com/TiyaAnlite/FocotServicesCommon/tracex"
	"testing"
)

func TestRedisHelper(t *testing.T) {
	cfg := &RedisConfig{}
	envx.MustLoadEnv(cfg)
	helper := RedisHelper{}
	err := helper.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	db := helper.DB()
	val, err := db.Get(context.Background(), "testStr").Result()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("testStr: %s", val)
}

func TestRedisHelperWithTrace(t *testing.T) {
	ctx := context.Background()
	traceCfg := &tracex.ServiceTraceHelper{}
	traceCfg.SetupTrace()
	defer traceCfg.Shutdown(ctx)
	tracer := traceCfg.NewTracer()

	cfg := &RedisConfig{}
	envx.MustLoadEnv(cfg)
	helper := RedisHelper{}
	err := helper.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	childCtx, span := tracer.Start(ctx, "key-get")
	defer span.End()
	db := helper.DB()
	val, err := db.Get(childCtx, "testStr").Result()
	if err != nil {
		t.Fatal(err.Error())
	}
	span.End()
	t.Logf("testStr: %s", val)
}
