# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости и git
RUN apk add --no-cache git

# Копируем только файлы модулей сначала для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы
COPY . .

# Генерируем документацию Swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/main.go -o ./docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o effective-mobile ./cmd/main.go

# Final stage
FROM alpine:3.18

WORKDIR /app

# Устанавливаем tzdata для работы с временными зонами
RUN apk add --no-cache tzdata

# Копируем бинарник и необходимые файлы
COPY --from=builder /app/effective-mobile .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/config/config.yaml ./config/config.yaml
COPY --from=builder /app/migrations ./migrations

# Настройки времени
ENV TZ=Europe/Moscow

# Указываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./effective-mobile"]