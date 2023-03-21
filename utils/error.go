package utils

import (
	"errors"
	"k8s.io/klog/v2"
)

func LogError(err error) {
	klog.ErrorDepth(1, err)
}

func LogAndReturnError(err error) error {
	klog.ErrorDepth(1, err)
	return err
}

func RecoverAndLogError() {
	if err := recover(); err != nil {
		klog.ErrorDepth(3, err)
	}
}

func RecoverAndLogAndWriteError(errVar *error) {
	err := recover()
	if err != nil {
		klog.ErrorDepth(3, err)
		switch e := err.(type) {
		case string:
			*errVar = errors.New(e)
		case error:
			*errVar = e
		}
	}
}
