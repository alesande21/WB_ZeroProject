package main

import (
	"AvitoProject/internal/app"
)

func main() {

	app.Run()

}

/*
host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable
POSTGRES_CONN=host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable;POSTGRES_DATABASE=postgres;POSTGRES_HOST=127.0.0.1;POSTGRES_PASSWORD=postgres;POSTGRES_PORT=5432;POSTGRES_USERNAME=postgres;SERVER_ADDRESS=127.0.0.1:8080
*/
