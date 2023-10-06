package dbx

import (
	"context"
	"github.com/TiyaAnlite/FocotServicesCommon/envx"
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
		t.Fatal(err)
	}
	t.Logf("testStr: %s", val)
}
