# syntax=docker/dockerfile:1

# Этап сборки
FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарник с именем main
RUN go build -o main .

# Финальный образ
FROM debian:bookworm-slim

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]
