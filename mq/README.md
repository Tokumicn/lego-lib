## 

```go
package main

import (
    "xxx.com/infrastructure/kkkk/configor"
    "xxx.com/infrastructure/kkkk/broker"
)

type Config struct {
    Kafka  broker.Config
    Rocket broker.Config
}

var (
    conf Config
)

func init() {
    if err := configor.Load("./configs/conf.toml", &conf); err != nil {
        panic(err)
    }
}

func main() {
    // 生产
    producer := broker.NewSyncProducer(broker.KAFKA, conf.Kafka)
    //producer := broker.NewSyncProducer(broker.ROCKET, conf.Rocket)
    producer.Send(ctx, "A", &broker.Message{Tag: "", Key: "", Value: []byte("what'up test")})

    // 消费
    consumer := broker.NewConsumer(broker.KAFKA, conf.Kafka)
    //consumer := broker.NewConsumer(broker.ROCKET, conf.Rocket)
    consumer.Recv("A", h1)
    consumer.Recv("B", h2)
    consumer.Start()
}

func h1(ctx context.Context, event broker.Event) error {
    fmt.Println(tracing.GetTraceID(ctx), event.GetTopic(), event.GetMessage().Tag, string(event.GetMessage().Value))
    return nil
}

func h2(ctx context.Context, event broker.Event) error {
    fmt.Println(tracing.GetTraceID(ctx), event.GetTopic(), event.GetMessage().Tag, string(event.GetMessage().Value))
    return nil
}
```

```toml
#conf.toml
[kafka]
    endpoints = ["localhost:9092"]
    group     = "feed"
    [[kafka.topics]]
        name     = "A"
        topic    = "test"
    [[kafka.topics]]
        name     = "B"
        topic    = "hello"

#[rocket]
#   endpoints  = ["http://MQ_xxx.cn-beijing.internal.aliyuncs.com:8080"]
#   group      = "GID-xx-waitdeposit-test-local2"
#   access_key = "accessxxxx"
#   secret_key = "secretyyyy"
#   instance   = "Minstance"
#   [[rocket.topics]]
#       name     = "A"
#       topic    = "xx-testing-local"
```
