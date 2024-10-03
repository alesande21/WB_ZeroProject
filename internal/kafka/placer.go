package kafka

import (
	config2 "WB_ZeroProject/internal/config"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type OrderPlacer struct {
	producer   *kafka.Producer
	topic      string
	deliveryCh chan kafka.Event
}

func NewOrderPlacer(conf *config2.ConfigKafka) (*OrderPlacer, error) {
	configMap := kafka.ConfigMap{
		"bootstrap.servers": conf.URL,
		"client.id":         "orderPlacer",
		"acks":              "all",
	}

	client, err := kafka.NewProducer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("kafka.NewProducer %w", err)
	}

	return &OrderPlacer{
		producer:   client,
		topic:      conf.Topic,
		deliveryCh: make(chan kafka.Event, 10000),
	}, nil
}

func (op *OrderPlacer) PlaceOrder(orderType string, size int) error {
	var (
		format  = fmt.Sprintf("%s - %d", orderType, size)
		payload = []byte(format)
	)

	err := op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &op.topic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
	},
		op.deliveryCh,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	<-op.deliveryCh

	return nil
}

func (op *OrderPlacer) GetOrder() error {
	op.producer.
	return nil
}

func (op *OrderPlacer) Close() {
	op.producer.Close()
}