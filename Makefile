all: build run

install:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.0.0

install_work_v:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

ver:
	oapi-codegen -version

genServer:
	oapi-codegen --config=configs/oapi/configServer.yaml api/ordersAPI.yml

build:
	#go build ./cmd/app/
	go build -o ./bin/app ./cmd/app
#	go build ./src/cmd/app/
	#go build -o build/server ./server

run:
	./app

init:
	go mod init "AvitoProject"

tidy:
	go mod tidy

test:
	go test .\internal\service\tender_test.go

cover:
	go test .\internal\service\tender_test.go -cover

test_full:
	go test ./...

cover_full:
	go test ./... -cover

cover_with_text:
	go test ./... -coverprofile=coverage.txt
	go tool cover -html coverage.txt -o index.html

.PHONY: all, install, ver, genServer, build, run, init, tidy, test, cover, test_full, cover_full, cover_with_text