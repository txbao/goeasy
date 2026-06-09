package mq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/txbao/goeasy/config"
)

type nsqMQ struct {
	cfg       config.MQ
	producer  *nsq.Producer
	consumers []*nsq.Consumer
	mu        sync.Mutex
}

func openNSQ(cfg config.MQ) (*nsqMQ, error) {
	nsqd := cfg.NSQDAddress()
	if nsqd == "" {
		return nil, errors.New("mq enabled but nsqd_addr/addr is empty")
	}
	ncfg := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqd, ncfg)
	if err != nil {
		return nil, fmt.Errorf("nsq producer: %w", err)
	}
	return &nsqMQ{cfg: cfg, producer: producer}, nil
}

func (n *nsqMQ) Publish(ctx context.Context, topic string, body []byte) error {
	_ = ctx
	if topic == "" {
		return errors.New("mq publish: topic is empty")
	}
	return n.producer.Publish(topic, body)
}

func (n *nsqMQ) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, body []byte) error) error {
	_ = ctx
	lookupd := n.cfg.LookupdAddress()
	if lookupd == "" {
		return errors.New("mq subscribe: lookupd_addr/addr is empty")
	}
	if topic == "" {
		return errors.New("mq subscribe: topic is empty")
	}
	ncfg := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, n.cfg.ConsumerChannel(), ncfg)
	if err != nil {
		return fmt.Errorf("nsq consumer: %w", err)
	}
	consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {
		hctx := context.Background()
		if err := handler(hctx, msg.Body); err != nil {
			return err
		}
		return nil
	}))
	if err := consumer.ConnectToNSQLookupd(lookupd); err != nil {
		return fmt.Errorf("nsq connect lookupd: %w", err)
	}
	n.mu.Lock()
	n.consumers = append(n.consumers, consumer)
	n.mu.Unlock()
	return nil
}

func (n *nsqMQ) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.producer != nil {
		n.producer.Stop()
		n.producer = nil
	}
	for _, c := range n.consumers {
		c.Stop()
	}
	n.consumers = nil
	return nil
}
