package envx

import (
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
