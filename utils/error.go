package utils

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/klog/v2"
)

func LogError(err error) {
	klog.ErrorDepth(1, err)
}

func LogAndReturnError(err error) error {
	klog.ErrorDepth(1, err)
	return err
}

func LogAndReturnErrorWithCtx(ctx context.Context, err error) error {
	ErrorWithCtx(ctx, err.Error(), 1)
	return err
}

func RecoverAndLogError() {
	if err := recover(); err != nil {
		klog.ErrorDepth(3, err)
	}
}

func RecoverAndLogErrorWithCtx(ctx context.Context) {
	if err := recover(); err != nil {
		switch e := err.(type) {
		case error:
			ErrorWithCtx(ctx, e.Error(), 3)
		case string:
			ErrorWithCtx(ctx, e, 3)
		default:
			ErrorWithCtx(ctx, fmt.Sprintf("%+v", e), 3)
		}
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

func RecoverAndLogAndWriteErrorWithCtx(ctx context.Context, errVar *error) {
	err := recover()
	if err != nil {
		switch e := err.(type) {
		case error:
			ErrorWithCtx(ctx, e.Error(), 3)
			*errVar = e
		case string:
			ErrorWithCtx(ctx, e, 3)
			*errVar = errors.New(e)
		default:
			msg := fmt.Sprintf("%+v", e)
			ErrorWithCtx(ctx, msg, 3)
			*errVar = errors.New(msg)
		}
	}
}
