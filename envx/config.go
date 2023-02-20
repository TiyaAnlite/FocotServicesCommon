package envx

import (
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
	"os"
)

func ReadYamlConfig(v interface{}, fileName ...string) error {
	file := "config.yaml"
	if len(fileName) > 0 {
		file = fileName[0]
	}
	dataBytes, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("Read config.yaml failed: %s", err.Error())
		return err
	}
	err = yaml.Unmarshal(dataBytes, v)
	if err != nil {
		klog.Fatalf("Failed parse config: %s", err.Error())
		return err
	}
	return nil
}

func MustReadYamlConfig(v interface{}, fileName ...string) {
	if err := ReadYamlConfig(v, fileName...); err != nil {
		klog.Fatal(err)
	}
}
