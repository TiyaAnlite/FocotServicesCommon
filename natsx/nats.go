package natsx

import (
	"github.com/nats-io/nats.go"
	"k8s.io/klog/v2"
	"time"
)

type NatsConfig struct {
	NatsUrl  string `json:"nats_url" yaml:"nats_url" env:"NATS_URL" envDefault:"127.0.0.1"`
	NatsName string `json:"nats_name" yaml:"nats_name" env:"NATS_NAME"`
	NatsNkey string `json:"nats_nkey" yaml:"nats_nkey" env:"NATS_NKEY"`
}

type NatsHelper struct {
	Nc   *nats.Conn
	Ec   *nats.EncodedConn
	Js   nats.JetStreamContext
	subs []*nats.Subscription

	// 普通消息发送
	Publish     func(subject string, data []byte) error
	PublishJson func(subject string, v interface{}) error
	// 请求应答
	Request     func(subj string, data []byte, timeout time.Duration) (*nats.Msg, error)
	RequestJson func(subject string, v interface{}, vPtr interface{}, timeout time.Duration) error
}

func (helper *NatsHelper) Open(cfg NatsConfig) error {
	klog.V(1).Infof("connecting to nats: %s", cfg.NatsUrl)
	var opts []nats.Option
	if cfg.NatsName != "" {
		opts = append(opts, nats.Name(cfg.NatsName))
	}
	if cfg.NatsNkey != "" {
		nkey, err := nats.NkeyOptionFromSeed(cfg.NatsNkey)
		if err != nil {
			return err
		}
		opts = append(opts, nkey)
	}
	opts = append(opts,
		nats.NoEcho(),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			if err != nil {
				klog.Errorf("nats disconnect: %v", err.Error())
			}
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			klog.Infof("nats reconnected: %s", c.ConnectedUrl())
		}),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			if s != nil {
				klog.Errorf("nats error in %q/%q: %v", s.Subject, s.Queue, err)
			} else {
				klog.Errorf("nats error: %v", err)
			}
		}),
	)

	var err error
	helper.Nc, err = nats.Connect(cfg.NatsUrl, opts...)
	if err != nil {
		return err
	}
	return helper.onConnected()
}

func (helper *NatsHelper) onConnected() error {
	helper.Publish = helper.Nc.Publish
	helper.Request = helper.Nc.Request
	var err error
	helper.Ec, err = nats.NewEncodedConn(helper.Nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}
	helper.PublishJson = helper.Ec.Publish
	helper.RequestJson = helper.Ec.Request
	helper.Js, err = helper.Nc.JetStream()
	if err != nil {
		return err
	}
	return err
}

func (helper *NatsHelper) Close() {
	helper.unsubscribe()
	if helper.Nc != nil && helper.Nc.IsConnected() {
		if err := helper.Nc.Drain(); err != nil {
			klog.Errorf("failed to drain: %v", err)
		}
		helper.Nc.Close()
	}
}

// AddNatsHandler 添加消息处理器
// 相当于调用连接的Subscribe，主要是多了一个自动Unsubscribe
func (helper *NatsHelper) AddNatsHandler(subject string, handler nats.MsgHandler) error {
	sub, err := helper.Nc.Subscribe(subject, handler)
	if err != nil {
		klog.Errorf("failed to subscribe to %s: %v", subject, err.Error())
		return err
	}
	helper.subs = append(helper.subs, sub)
	return nil
}

func (helper *NatsHelper) AddNatsJSONHandler(subject string, handler nats.Handler) error {
	sub, err := helper.Ec.Subscribe(subject, handler)
	if err != nil {
		klog.Errorf("failed to subscribe to %s: %v", subject, err.Error())
		return err
	}
	helper.subs = append(helper.subs, sub)
	return nil
}

// AddSubscribe 添加自定义的订阅主题，主要用于统一Unsubscribe
func (helper *NatsHelper) AddSubscribe(sub *nats.Subscription) {
	helper.subs = append(helper.subs, sub)
}

func (helper *NatsHelper) unsubscribe() {
	for _, sub := range helper.subs {
		if err := sub.Unsubscribe(); err != nil {
			klog.Errorf("failed to unsubscribe: %v", err)
		}
		if err := sub.Drain(); err != nil {
			klog.Errorf("failed to drain subscription: %v", err)
		}
	}
}
