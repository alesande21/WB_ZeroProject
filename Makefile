all: buildDefault runDefault

buildDefault:
	go build ./cmd/WB_ZeroProject/main.go

runDefault:
	./main.exe

init:
	go mod init "WB_ZeroProject"

tidy:
	go mod tidy