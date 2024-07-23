package dbx

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"k8s.io/klog/v2"
	"time"
)

type RedisConfig struct {
	Host      string `json:"host" env:"REDIS_HOST,required" envDefault:"localhost"`
	Port      int    `json:"port" env:"REDIS_PORT,required" envDefault:"6379"`
	Pass      string `json:"pass" env:"REDIS_PASS"`
	DB        int    `json:"db" env:"REDIS_DB"`
	Telemetry bool   `env:"REDIS_TELEMETRY" envDefault:"true"`
}

type RedisHelper struct {
	rdb *redis.Client
}

func (r *RedisHelper) Open(cfg *RedisConfig) error {
	if r.rdb != nil {
		klog.Warningf("RedisHelper opened, will close old connection")
		r.rdb.Close()
	}
	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       cfg.DB,
	}
	klog.Infof("connection to redis: %s:%d", cfg.Host, cfg.Port)
	r.rdb = redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if cfg.Telemetry {
		klog.Info("RedisHelper telemetry on")
		if err := redisotel.InstrumentTracing(r.rdb); err != nil {
			return fmt.Errorf("failed to instrument tracing: %s", err.Error())
		}
		if err := redisotel.InstrumentMetrics(r.rdb); err != nil {
			return fmt.Errorf("failed to instrument metrics: %s", err.Error())
		}
	}
	err := r.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %s", err.Error())
	}
	return nil
}

func (r *RedisHelper) DB() *redis.Client {
	return r.rdb
}
