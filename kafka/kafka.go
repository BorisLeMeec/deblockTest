package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"

	"deblockTest/pkg"
)

type Kafka struct {
	config *Config
	writer *kafka.Writer
}

func NewFromConfig(cfg *Config) *Kafka {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Broker),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Kafka{config: cfg, writer: writer}
}

func (k *Kafka) Publish(ctx context.Context, msgs []pkg.TxMessage) {
	kafkaMsgs := make([]kafka.Message, len(msgs))
	for i, m := range msgs {
		value, _ := json.Marshal(m)
		kafkaMsgs[i] = kafka.Message{Value: value}
	}

	for attempts := 0; attempts < 10; attempts++ {
		if err := k.writer.WriteMessages(ctx, kafkaMsgs...); err == nil {
			break
		} else {
			log.Printf("Kafka write failed (attempt %d): %v", attempts+1, err)
			time.Sleep(time.Second << attempts)
		}
		if attempts == 9 {
			log.Fatal("Failed to write to Kafka after retries")
		}
	}
}

func (k *Kafka) Close() {
	k.writer.Close()
}
