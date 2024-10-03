package kafka

import (
	config2 "WB_ZeroProject/internal/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type orderConsumer struct {
	Consumer *kafka.Consumer
}

func NewOrderConsumer(conf *config2.ConfigKafka) (*OrderPlacer, error) {
	configMap := kafka.ConfigMap{
		"bootstrap.servers":  conf.URL,
		"group.id":           conf.Topic,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	client, err := kafka.
	return &OrderPlacer{
		producer:   p,
		topic:      topic,
		deliveryCh: make(chan kafka.Event, 10000),
	}
}
