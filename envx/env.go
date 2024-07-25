package envx

import (
	"context"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"k8s.io/klog/v2"
	"strings"
	"sync"
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

type rdbEnvLoaderOptions struct {
	envOptions          env.Options
	autoLoadProjectName string
	notifyMq            *nats.Conn
	pendingLock         *sync.RWMutex
	ErrorAtNotFound     bool
	loadTimeout         time.Duration
}

type RdbEnvLoaderOption func(options *rdbEnvLoaderOptions)

func WithRdbEnvCustomOptions(opt env.Options) RdbEnvLoaderOption {
	return func(options *rdbEnvLoaderOptions) {
		options.envOptions = opt
	}
}

func WithRdbEnvAutoLoad(projectName string, mq *nats.Conn, lock *sync.RWMutex) RdbEnvLoaderOption {
	return func(options *rdbEnvLoaderOptions) {
		options.autoLoadProjectName = projectName
		options.notifyMq = mq
		options.pendingLock = lock
	}
}

func WithRdbEnvErrorAtNotFound(at bool) RdbEnvLoaderOption {
	return func(options *rdbEnvLoaderOptions) {
		options.ErrorAtNotFound = at
	}
}

func WithRdbEnvLoadTimeout(timeout time.Duration) RdbEnvLoaderOption {
	return func(options *rdbEnvLoaderOptions) {
		options.loadTimeout = timeout
	}
}

func LoadEnvFromRedis(v any, r *redis.Client, key string, option ...RdbEnvLoaderOption) error {
	opt := &rdbEnvLoaderOptions{loadTimeout: time.Second * 3}
	for _, o := range option {
		o(opt)
	}
	ctx, cancel := context.WithTimeout(context.Background(), opt.loadTimeout)
	defer cancel()
	val, err := r.HGetAll(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		if opt.ErrorAtNotFound {
			return fmt.Errorf("LoadEnvFromRedis: key[%s] not found", key)
		}
		klog.Warningf("LoadEnvFromRedis: key[%s] not found", key)
	}
	if err != nil {
		return err
	}
	opt.envOptions.Environment = val
	err = LoadEnv(v, opt.envOptions)
	if err != nil {
		return err
	}
	// auto load
	if opt.autoLoadProjectName != "" && opt.notifyMq != nil && opt.pendingLock != nil {
		subjectPrefix := fmt.Sprintf("%s.%s", "envAutoLoad", opt.autoLoadProjectName)
		handler := envAutoReloadHandler(v, key, r, subjectPrefix, opt)
		if _, err := opt.notifyMq.Subscribe(subjectPrefix, handler); err != nil {
			return fmt.Errorf("failed to subscribe to envAutoLoad subject [%s]: %s", opt.autoLoadProjectName, err.Error())
		}
		if _, err := opt.notifyMq.Subscribe(subjectPrefix+".>", handler); err != nil {
			return fmt.Errorf("failed to subscribe to envAutoLoad subject [%s]: %s", opt.autoLoadProjectName+".>", err.Error())
		}
		klog.Infof("LoadEnvFromRedis: setup autoload at: %s", subjectPrefix)
	}
	return nil
}

func MustLoadEnvFromRedis(v any, r *redis.Client, key string, option ...RdbEnvLoaderOption) {
	if err := LoadEnvFromRedis(v, r, key, option...); err != nil {
		klog.Fatal(err)
	}
}

func envAutoReloadHandler(v any, key string, r *redis.Client, subjectPrefix string, opt *rdbEnvLoaderOptions) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		opt.pendingLock.Lock()
		defer opt.pendingLock.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), opt.loadTimeout)
		defer cancel()
		subject := strings.TrimPrefix(msg.Subject, subjectPrefix)
		if subject == "" {
			// load all
			klog.Info("[AutoRedisEnv]Auto reloading all config")
			val, err := r.HGetAll(ctx, key).Result()
			if errors.Is(err, redis.Nil) {
				errMsg := fmt.Sprintf("LoadEnvFromRedis: key[%s] not found", key)
				klog.Errorf("[AutoRedisEnv]%s", errMsg)
				_ = msg.Respond([]byte(errMsg))
				return
			}
			if err != nil {
				klog.Errorf("[AutoRedisEnv]%s", err.Error())
				_ = msg.Respond([]byte(err.Error()))
				return
			}
			opt.envOptions.Environment = val
			if err := LoadEnv(v, opt.envOptions); err != nil {
				klog.Errorf("[AutoRedisEnv]%s", err.Error())
				_ = msg.Respond([]byte(err.Error()))
				return
			}
			_ = msg.Respond([]byte("ok"))
			return
		}
		subject = strings.TrimPrefix(subject, ".")
		// load single
		klog.Infof("[AutoRedisEnv]Auto reloading config: %s", subject)
		val, err := r.HGet(ctx, key, subject).Result()
		if errors.Is(err, redis.Nil) {
			errMsg := fmt.Sprintf("LoadEnvFromRedis: key[%s]->field[%s] not found", key, subject)
			klog.Errorf("[AutoRedisEnv]%s", errMsg)
			_ = msg.Respond([]byte(errMsg))
			return
		}
		if err != nil {
			klog.Errorf("[AutoRedisEnv]%s", err.Error())
			_ = msg.Respond([]byte(err.Error()))
			return
		}
		opt.envOptions.Environment[subject] = val
		if err := LoadEnv(v, opt.envOptions); err != nil {
			klog.Errorf("[AutoRedisEnv]%s", err.Error())
			_ = msg.Respond([]byte(err.Error()))
			return
		}
		_ = msg.Respond([]byte("ok"))
		return
	}
}
