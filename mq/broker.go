package mq

import (
	"context"
	"errors"
)

const (
	KAFKA  = 1
	ROCKET = 2
)

// CallbackHandler 消费回调函数
type CallbackHandler func(ctx context.Context, event Event) error

// Event 回调参数接口
type Event interface {
	GetTopic() string
	GetMessage() Message
}

// Message 消息结构
type Message struct {
	Tag   string
	Key   string
	Value []byte
}

// Consumer 消费接口
type Consumer interface {
	Recv(topicName string, h CallbackHandler) error
	Start() error
	Close() error
}

// NewConsumer 创建Consumer
func NewConsumer(broker int, conf *Config) Consumer {
	var (
		consumer Consumer
		err      error
	)

	switch broker {
	case KAFKA:
		consumer, err = NewKafkaConsumer(conf)
	default:
		consumer, err = nil, errors.New("unknow broker type")
	}

	if err != nil {
		panic(err)
	}
	return consumer
}

// SyncProducer 同步生产接口
type SyncProducer interface {
	Send(ctx context.Context, topicName string, msg *Message) error
	Close() error
}

// NewSyncProducer 创建SyncProducer
func NewSyncProducer(broker int, conf *Config) SyncProducer {
	var (
		producer SyncProducer
		err      error
	)

	switch broker {
	case KAFKA:
		producer, err = NewKafkaSyncProducer(conf)
	default:
		producer, err = nil, errors.New("unknow broker type")
	}

	if err != nil {
		panic(err)
	}
	return producer
}

// AsyncProducer 异步生产接口
type AsyncProducer interface {
	Send(ctx context.Context, topicName string, msg *Message) error
	Close() error
}

// NewAsyncProducer 创建AsyncProducer
func NewAsyncProducer(broker int, conf Config) AsyncProducer {
	var (
		producer AsyncProducer
		err      error
	)

	switch broker {
	case KAFKA:
		producer, err = NewKafkaAsyncProducer(conf)
	default:
		producer, err = nil, errors.New("unknow broker type")
	}

	if err != nil {
		panic(err)
	}
	return producer
}

// TopicConfig topic配置
type TopicConfig struct {
	Name  string `toml:"name"`
	Topic string `toml:"topic"`
}

// Config 消息队列配置项
type Config struct {
	Broker    string        `toml:"broker"`
	Endpoints []string      `toml:"endpoints"`
	AccessKey string        `toml:"access_key"`
	SecretKey string        `toml:"secret_key"`
	Instance  string        `toml:"instance"`
	Group     string        `toml:"group"`
	Topics    []TopicConfig `toml:"topics"`
}
