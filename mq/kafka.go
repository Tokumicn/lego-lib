package mq

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tokumicn/lego-lib/logs"
	"runtime/debug"
	"time"

	"github.com/Shopify/sarama"
)

// KafkaConsumer kafka消费者结构
type KafkaConsumer struct {
	client sarama.ConsumerGroup
	topics map[string]string
}

// NewKafkaConsumer 创建KafkaConsumer
func NewKafkaConsumer(conf *Config) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	config.ChannelBufferSize = 128

	client, err := sarama.NewConsumerGroup(conf.Endpoints, conf.Group, config)
	if err != nil {
		return nil, err
	}

	consumer := &KafkaConsumer{
		client: client,
		topics: make(map[string]string),
	}

	for _, tc := range conf.Topics {
		consumer.topics[tc.Name] = tc.Topic
	}
	return consumer, nil
}

// Recv 消费消息 设置回调函数
func (c *KafkaConsumer) Recv(name string, callback CallbackHandler) error {
	topic, ok := c.topics[name]
	if !ok {
		return errors.New("kafka consume find topic failed")
	}

	go func() {
		for {
			select {
			case err := <-c.client.Errors():
				if err != nil {
					fmt.Printf("kafka consume recv err:%v", err)
				}
			default:
				if err := c.client.Consume(context.Background(), []string{topic}, &GroupHandler{
					TopicName:       name,
					CallbackHandler: callback,
					ConsumerGroup:   c.client,
				}); err != nil {
					fmt.Printf("kafka consume invoke err:%v", err)
				}
			}
		}
	}()
	return nil
}

// Start 启动消费消息
func (c *KafkaConsumer) Start() error {
	return nil
}

// Close 关闭消费
func (c *KafkaConsumer) Close() error {
	return c.client.Close()
}

// KafkaSyncProducer kafka同步生产者结构
type KafkaSyncProducer struct {
	client sarama.SyncProducer
	topics map[string]string
}

// NewKafkaSyncProducer 创建KafkaSyncProducer
func NewKafkaSyncProducer(conf *Config) (*KafkaSyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	client, err := sarama.NewSyncProducer(conf.Endpoints, config)
	if err != nil {
		return nil, err
	}

	producer := &KafkaSyncProducer{
		client: client,
		topics: make(map[string]string),
	}

	for _, tc := range conf.Topics {
		producer.topics[tc.Name] = tc.Topic
	}

	return producer, nil
}

// Send 同步发送消息
func (p *KafkaSyncProducer) Send(ctx context.Context, name string, msg *Message) error {
	if p.client == nil {
		return errors.New("kafka sync broker client nil")
	}

	topic, ok := p.topics[name]
	if !ok {
		return errors.New("kafka sync find topic failed")
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg.Value),
	}
	if len(msg.Key) != 0 {
		message.Key = sarama.ByteEncoder(msg.Key)
	}

	_, _, err := p.client.SendMessage(message)
	return err
}

// Close 关闭同步生产者
func (p *KafkaSyncProducer) Close() error {
	return p.client.Close()
}

// KafkaAsyncProducer kafka异步生产者结构
type KafkaAsyncProducer struct {
	client sarama.AsyncProducer
	topics map[string]string
}

// NewKafkaAsyncProducer 创建KafkaAsyncProducer
func NewKafkaAsyncProducer(conf Config) (*KafkaAsyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	client, err := sarama.NewAsyncProducer(conf.Endpoints, config)
	if err != nil {
		return nil, err
	}

	producer := &KafkaAsyncProducer{
		client: client,
		topics: make(map[string]string),
	}

	for _, tc := range conf.Topics {
		producer.topics[tc.Name] = tc.Topic
	}

	go producer.asyncRecvErrors()
	return producer, nil
}

// Send 异步发送消息
func (p *KafkaAsyncProducer) Send(ctx context.Context, name string, msg *Message) error {
	if p.client == nil {
		return errors.New("kafka async broker client nil")
	}

	topic, ok := p.topics[name]
	if !ok {
		return errors.New("kafka async find topic failed")
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg.Value),
	}
	if len(msg.Key) != 0 {
		message.Key = sarama.ByteEncoder(msg.Key)
	}

	p.client.Input() <- message
	return nil
}

// Close 关闭异步生产者
func (p *KafkaAsyncProducer) Close() error {
	return p.client.Close()
}

func (p *KafkaAsyncProducer) asyncRecvErrors() {
	if p.client == nil {
		return
	}

	for err := range p.client.Errors() {
		fmt.Printf("kafka async recv err:%v", err)
	}
}

// GroupHandler 回调封装 sarama约定
type GroupHandler struct {
	TopicName       string
	CallbackHandler CallbackHandler
	ConsumerGroup   sarama.ConsumerGroup
}

// Setup ...
func (*GroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup ...
func (*GroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 回调函数执行
func (g *GroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer doRecover()
	for msg := range claim.Messages() {
		if err := g.CallbackHandler(context.TODO(), &KafkaEvent{
			Topic:   msg.Topic,
			Message: Message{Tag: getMessageTag(msg.Headers), Key: string(msg.Key), Value: msg.Value},
		}); err == nil {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}

func doRecover() {
	if r := recover(); r != nil {
		logs.Errorf("[PANIC] time:%d err:%v stack", time.Now(), r, string(debug.Stack()))
	}
}

// KafkaEvent kafka消息事件
type KafkaEvent struct {
	Topic   string
	Message Message
}

// GetTopic 获取事件对应的Topic
func (k *KafkaEvent) GetTopic() string {
	return k.Topic
}

// GetMessage 获取事件对应的Message
func (k *KafkaEvent) GetMessage() Message {
	return k.Message
}

func getMessageTag(headers []*sarama.RecordHeader) string {
	for _, header := range headers {
		if string(header.Key) == "TAGS" {
			return string(header.Value)
		}
	}
	return ""
}
