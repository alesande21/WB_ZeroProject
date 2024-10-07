# Используем официальный образ Go в качестве базового образа для сборки
FROM golang:1.23.1-bookworm AS builder

# Устанавливаем рабочую директорию для сборки
WORKDIR /build/

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Сборка приложения
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags "-extldflags -static" \
      -o wb_service ./cmd/consumer

# Финальный образ
FROM debian:bookworm-20240904-slim

# Устанавливаем необходимые пакеты
RUN set -x && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      ca-certificates \
      libc6 \
      libgcc-s1 \
      libgssapi-krb5-2 \
      librdkafka1 \
      && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /api/
ENV PATH=/api/bin/:$PATH

# Копируем скомпилированный бинарник из стадии сборки
COPY --from=builder /build/wb_service ./bin/wb_service
COPY --from=builder /build/env.example .


# Экспонируем порт
EXPOSE 8080

# Определяем точку входа
CMD ["./bin/wb_service"]
