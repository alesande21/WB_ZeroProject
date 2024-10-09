package kafka

import (
	entity2 "WB_ZeroProject/internal/entity"
	"WB_ZeroProject/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	log2 "github.com/sirupsen/logrus"
)

type OrderConsumer struct {
	Consumer     *kafka.Consumer
	Producer     *kafka.Producer
	Topic        string
	OrderService *service.OrderService
}

type eventGetResponse struct {
	Type          string        `json:"type"`
	Order         entity2.Order `json:"order"`
	Status        bool          `json:"status"`
	CorrelationID string        `json:"correlation_id"`
}

func NewOrderConsumer(conf *ConfigKafka, orderSerivce *service.OrderService, groupID string) (*OrderConsumer, error) {
	configMapConsumer := kafka.ConfigMap{
		"bootstrap.servers":  conf.URL,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	clientConsumer, err := kafka.NewConsumer(&configMapConsumer)
	if err != nil {
		return nil, fmt.Errorf("-> kafka.NewConsumer: ошибка при создании Consumer: %w", err)
	}

	err = clientConsumer.Subscribe("orders.event.request", nil)
	if err != nil {
		return nil, fmt.Errorf("-> clientConsumer.Subscribe: ошибка при подписке на топик %s: %w", conf.Topic, err)
	}

	configMapProducer := kafka.ConfigMap{
		"bootstrap.servers": conf.URL,
		"client.id":         "orderConsumer",
		"acks":              "all",
	}

	clientProducer, err := kafka.NewProducer(&configMapProducer)
	if err != nil {
		errClose := clientConsumer.Close()
		if errClose != nil {
			return nil, fmt.Errorf("-> kafka.NewProducer-> clientConsumer.Close: ошибка при создании Producer: %w. Ошибка при закрытии Consumer: %w", err, errClose)
		}
		return nil, fmt.Errorf("-> kafka.NewProducer: ошибка при создании Producer: %w", err)
	}

	return &OrderConsumer{
		Consumer:     clientConsumer,
		Producer:     clientProducer,
		Topic:        conf.Topic,
		OrderService: orderSerivce,
	}, nil
}

func (oc *OrderConsumer) ListenAndServe(ctx context.Context) {
	commit := func(msg *kafka.Message) {
		if _, err := oc.Consumer.CommitMessage(msg); err != nil {
			log2.Errorf("ListenAndServe-> op.consumer.CommitMessage: коммит не выполнен: %s", err)
		}
	}

	run := true

	for run {
		select {
		case <-ctx.Done():
			log2.Info("Обработчик заказов остановлен...")
			run = false
			break

		default:
			msg, ok := oc.Consumer.Poll(150).(*kafka.Message)
			if !ok {
				continue
			}

			var evt event
			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&evt); err != nil {
				log2.Errorf("ListenAndServe-> json.NewDecoder: ошибка при декодировании event: %s", err)
				commit(msg)
				continue
			}

			//ok = false

			switch evt.Type {
			case "orders.event.request.create":
				var createEvent eventCreate
				if err := json.Unmarshal(msg.Value, &createEvent); err != nil {
					log2.Errorf("ListenAndServe-> json.NewDecoder: ошибка при декодировании eventCreate: %s", err)
					commit(msg)
					continue
				}

				orderIDs, err := oc.OrderService.Repo.CreateOrder(ctx, createEvent.Value)
				if err == nil {
					log2.Infof("Заказы успешно добвлены: %v", orderIDs)
					commit(msg)
				} else {
					log2.Errorf("ListenAndServe-> oc.OrderService.Repo.CreateOrder%s", err.Error())
				}

			case "orders.event.request.getByID":

				var getEvent eventGet
				if err := json.Unmarshal(msg.Value, &getEvent); err != nil {
					log2.Errorf("ListenAndServe-> json.NewDecoder: ошибка при декодировании getEvent: %s", err)
					commit(msg)
					continue
				}

				order, err := oc.OrderService.GetOrderById(ctx, getEvent.Value)
				responseEvent := eventGetResponse{
					Type:          "orders.event.response",
					Status:        true,
					CorrelationID: getEvent.CorrelationID,
				}
				if err != nil {
					log2.Errorf("ListenAndServe-> oc.OrderService.GetOrderById%s", err.Error())
					responseEvent.Status = false
				} else {
					responseEvent.Order = *order
				}

				var b bytes.Buffer
				err = json.NewEncoder(&b).Encode(responseEvent)
				if err != nil {
					log2.Errorf("ListenAndServe-> json.NewDecoder.Encode: ошибка при кодировании responseEvent: %s", err)
					continue
				}

				msgResp := kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &oc.Topic,
						Partition: kafka.PartitionAny,
					},
					Value: b.Bytes(),
				}

				err = oc.Producer.Produce(&msgResp, nil)
				if err == nil {
					log2.Infof("ListenAndServe: ответ с заказом %s отправлен.", getEvent.Value)
				} else {
					log2.Errorf("ListenAndServe-> oc.Producer.Produce: ошибка при отправке ответа с заказом %s: %s", getEvent.Value, err)
				}
				commit(msg)
			default:
				log2.Infof("Неизвестный тип события: %s", evt.Type)
				commit(msg)
			}

		}
	}
}

func (oc *OrderConsumer) Close() error {
	oc.Producer.Close()
	return oc.Consumer.Close()
}
