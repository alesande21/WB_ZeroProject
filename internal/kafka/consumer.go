package kafka

import (
	config2 "WB_ZeroProject/internal/config"
	"WB_ZeroProject/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type OrderConsumer struct {
	Consumer     *kafka.Consumer
	OrderService *service.OrderService
}

func NewOrderConsumer(conf *config2.ConfigKafka, orderSerivce *service.OrderService, groupID string) (*OrderConsumer, error) {
	configMap := kafka.ConfigMap{
		"bootstrap.servers":  conf.URL,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	client, err := kafka.NewConsumer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании -> kafka.NewConsumer %w", err)
	}

	err = client.Subscribe(conf.Topic+"*", nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подписке -> client.Subscribe %w", err)
	}

	return &OrderConsumer{
		Consumer:     client,
		OrderService: orderSerivce,
	}, nil
}

func (oc *OrderConsumer) ListenAndServe(ctx context.Context) {
	commit := func(msg *kafka.Message) {
		if _, err := oc.Consumer.CommitMessage(msg); err != nil {
			log.Printf("Коммит провален: %s", err)
		}
	}

	run := true

	for run {
		select {
		case <-ctx.Done():
			log.Printf("Обработчик заказов остановлен...")
			run = false
			break
		default:
			msg, ok := oc.Consumer.Poll(150).(*kafka.Message)
			if !ok {
				continue
			}

			var evt eventGet

			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&evt); err != nil {
				log.Printf("Ошибка при декодировании eventGet: %s", err)

				commit(msg)

				continue
			}

			ok = false

			switch evt.Type {
			case "orders.event.create":

			case "orders.event.getbyID":

			}

		}

	}
}

func (oc *OrderConsumer) Close() error {
	return oc.Consumer.Close()
}
