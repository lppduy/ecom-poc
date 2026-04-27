package event

import (
	"context"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
	}
	return &KafkaPublisher{writer: w}
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic, key, value string) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: []byte(value),
	})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
