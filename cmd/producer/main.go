package main

import (
	"WB_ZeroProject/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {

	err := app.RunProducer()
	if err != nil {
		logrus.Errorf("app.RunProducer%s", err.Error())
	}

}

/*
host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable
POSTGRES_CONN=host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable;POSTGRES_DATABASE=postgres;POSTGRES_HOST=127.0.0.1;POSTGRES_PASSWORD=postgres;POSTGRES_PORT=5432;POSTGRES_USERNAME=postgres;SERVER_ADDRESS=127.0.0.1:8080

KAFKA_HOST=127.0.0.1;KAFKA_PORT=9092;KAFKA_TOPIC=orders;KAFKA_CONN=localhost:9092;POSTGRES_CONN=host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable;POSTGRES_DATABASE=postgres;POSTGRES_HOST=127.0.0.1;POSTGRES_PASSWORD=postgres;POSTGRES_PORT=5432;POSTGRES_USERNAME=postgres;SERVER_ADDRESS=localhost:8080
*/
