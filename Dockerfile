#FROM golang:1.23.1-alpine3.20 AS build
#COPY . /home/src
#WORKDIR /home/src
#RUN apk add make && make build
#
#FROM alpine:3.20
#EXPOSE 8080
#WORKDIR /app
#COPY --from=build /home/src/bin/app /app/avito_service
#
#ENTRYPOINT [ "/bin/sh", "-c", "/app/avito_service" ]

# Используем официальный образ Go в качестве базового образа для сборки
FROM golang:1.22 AS builder
LABEL authors="alesande"

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o avito_service ./cmd/app

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates postgresql-client

# Создаем директорию для приложения
WORKDIR /app

# Копируем скомпилированный бинарник из стадии сборки
COPY --from=builder /app/avito_service /app/avito_service

# Делаем бинарный файл исполняемым
RUN chmod +x /app/avito_service

# Настройка порта, который будет использоваться
EXPOSE 8080

# Определяем точку входа
ENTRYPOINT ["/app/avito_service"]







