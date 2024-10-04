package kafka

import (
	config2 "WB_ZeroProject/internal/config"
	entity2 "WB_ZeroProject/internal/entity"
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
	Producer     *kafka.Producer
	OrderService *service.OrderService
}

type eventGetResponse struct {
	Type          string
	Order         entity2.Order
	Status        bool
	CorrelationID string
}

func NewOrderConsumer(conf *config2.ConfigKafka, orderSerivce *service.OrderService, groupID string) (*OrderConsumer, error) {
	configMapConsumer := kafka.ConfigMap{
		"bootstrap.servers":  conf.URL,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	clientConsumer, err := kafka.NewConsumer(&configMapConsumer)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании -> kafka.NewConsumer %w", err)
	}

	err = clientConsumer.Subscribe(conf.Topic+".event.request*", nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подписке -> client.Subscribe %w", err)
	}

	configMapProducer := kafka.ConfigMap{
		"bootstrap.servers": conf.URL,
		"client.id":         "orderConsumer",
		"acks":              "all",
	}

	clientProducer, err := kafka.NewProducer(&configMapProducer)
	if err != nil {
		err := clientConsumer.Close()
		if err != nil {
			return nil, fmt.Errorf("ошибка при закрытии -> clientConsumer.Close %w", err)
		}
		return nil, fmt.Errorf("ошибка при инициализации -> kafka.NewProducer %w", err)
	}

	return &OrderConsumer{
		Consumer:     clientConsumer,
		Producer:     clientProducer,
		OrderService: orderSerivce,
	}, nil
}

func (oc *OrderConsumer) ListenAndServe(ctx context.Context) {
	commit := func(msg *kafka.Message) {
		if _, err := oc.Consumer.CommitMessage(msg); err != nil {
			log.Printf("Коммит не выполнен: %s", err)
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

			var evt event

			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&evt); err != nil {
				log.Printf("Ошибка при декодировании event: %s", err)

				commit(msg)

				continue
			}

			//ok = false

			switch evt.Type {
			case "orders.event.request.create":
				var createEvent eventCreate
				if err := json.Unmarshal(evt.Value, &createEvent); err != nil {
					log.Printf("Ошибка при декодировании createEvent: %s", err)
					commit(msg)
					continue
				}
				orderIDs, err := oc.OrderService.Repo.CreateOrder(ctx, createEvent.Value)
				if err == nil {
					log.Println("Заказы успешно добвлены: ", orderIDs)
					commit(msg)
				} else {
					log.Println("Заказы не добавлены в базу данных: ", orderIDs)
				}

			case "orders.event.request.getByID":

				var getEvent eventGet
				if err := json.Unmarshal(evt.Value, &getEvent); err != nil {
					log.Printf("Ошибка при декодировании getEvent: %s", err)
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
					log.Printf("Заказ с ID %s не найден: %s", evt.Value, err)
					responseEvent.Status = false
				} else {
					responseEvent.Order = *order
				}

				var b bytes.Buffer
				err = json.NewEncoder(&b).Encode(responseEvent)
				if err != nil {
					log.Printf("Ошибка при кодировании ответа: %s", err)
					continue
				}

				msgResp := kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &responseEvent.Type,
						Partition: kafka.PartitionAny,
					},
					Value: b.Bytes(),
				}

				err = oc.Producer.Produce(&msgResp, nil)

				if err == nil {
					log.Printf("Ответ с заказом отправлен: %+v", order)
				} else {
					log.Printf("Ошибка при отправке ответа: %s", err)
				}
				commit(msg)
			default:
				log.Printf("Неизвестный тип события: %s", evt.Type)
				commit(msg)
			}

		}

	}
}

//func (oc *OrderConsumer) CreateOrderes(ctx context.Context) {
//
//}

func (oc *OrderConsumer) Close() error {
	oc.Producer.Close()
	return oc.Consumer.Close()
}
