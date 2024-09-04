all: buildDefault runDefault

install:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.0.0

install_work_v:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

ver:
	oapi-codegen -version

genServer:
	oapi-codegen --config=configs/oapi/configServer.yaml api/ordersAPI.yml

buildDefault:
	go build ./cmd/WB_ZeroProject/main.go

runDefault:
	./main.exe

init:
	go mod init "WB_ZeroProject"

tidy:
	go mod tidy