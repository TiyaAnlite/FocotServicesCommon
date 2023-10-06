package envx

import (
	"github.com/TiyaAnlite/FocotServicesCommon/dbx"
	"k8s.io/klog/v2"
	"testing"
)

type testEnv struct {
	A string `env:"a"`
	B string `env:"b,required"`
	C string `env:"c" envDefault:"c-default-string"`
}

func TestEnv(t *testing.T) {
	c := testEnv{}
	MustLoadEnv(&c)
	klog.Infof("a: %s, b: %s, c: %s", c.A, c.B, c.C)
}

func TestRedisEnv(t *testing.T) {
	c := testEnv{}
	cfg := &dbx.RedisConfig{}
	MustLoadEnv(cfg)
	helper := dbx.RedisHelper{}
	err := helper.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	db := helper.DB()
	MustLoadEnvFromRedis(&c, db, "config")
	klog.Infof("a: %s, b: %s, c: %s", c.A, c.B, c.C)
}
