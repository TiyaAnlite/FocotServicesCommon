package envx

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"k8s.io/klog/v2"
)

func LoadEnv(v interface{}, opts ...env.Options) error {
	_ = godotenv.Load()
	return env.Parse(v, opts...)
}

func MustLoadEnv(v interface{}, opts ...env.Options) {
	if err := LoadEnv(v, opts...); err != nil {
		klog.Fatal(err)
	}
}
