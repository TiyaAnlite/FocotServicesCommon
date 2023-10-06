package envx

import (
	"context"
	"github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	"k8s.io/klog/v2"
	"time"
)

func LoadEnv(v any, opts ...env.Options) error {
	return env.Parse(v, opts...)
}

func MustLoadEnv(v any, opts ...env.Options) {
	if err := LoadEnv(v, opts...); err != nil {
		klog.Fatal(err)
	}
}

func LoadEnvFromRedis(v any, r *redis.Client, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	val, err := r.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		klog.Warningf("LoadEnvFromRedis: key[%s] not found", key)
	}
	if err != nil {
		return err
	}
	return LoadEnv(v, env.Options{Environment: val})
}

func MustLoadEnvFromRedis(v any, r *redis.Client, key string) {
	if err := LoadEnvFromRedis(v, r, key); err != nil {
		klog.Fatal(err)
	}
}
