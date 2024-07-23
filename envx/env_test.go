package envx

import (
	"github.com/TiyaAnlite/FocotServicesCommon/dbx"
	"github.com/TiyaAnlite/FocotServicesCommon/natsx"
	"k8s.io/klog/v2"
	"sync"
	"testing"
	"time"
)

type testEnv struct {
	A string `env:"a"`
	B string `env:"b,required"`
	C string `env:"c" envDefault:"c-default-string"`
	D string `env:"a.b.c"`
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

func TestRedisAutoEnv(t *testing.T) {
	c := testEnv{}
	cfg := &dbx.RedisConfig{}
	MustLoadEnv(cfg)
	helper := dbx.RedisHelper{}
	err := helper.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	nats := natsx.NatsHelper{}
	natsCfg := &natsx.NatsConfig{}
	MustLoadEnv(natsCfg)
	if err := nats.Open(*natsCfg); err != nil {
		t.Fatal(err)
	}
	db := helper.DB()
	lock := &sync.RWMutex{}
	MustLoadEnvFromRedis(&c, db, "config", WithRdbEnvAutoLoad("testProj", nats.Nc, lock))
	for range time.Tick(time.Second) {
		lock.RLock()
		klog.Infof("a: %s, b: %s, c: %s, d: %s", c.A, c.B, c.C, c.D)
		time.Sleep(time.Millisecond * 500) // do something
		lock.RUnlock()
	}
}
