# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Копирование go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:3.19

WORKDIR /app

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

# Копирование бинарника из builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .

# Создание непривилегированного пользователя
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 9090

CMD ["./main", "-config", "config.yaml"]
