# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Переходим в директорию микросервиса и собираем
WORKDIR /app/cmd/file-storage
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o file-storage .

# Финальный этап
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/cmd/file-storage/file-storage .

# Копируем миграции
COPY --from=builder /app/migrations/file-storage ./migrations

# Создаем директорию для загрузок
RUN mkdir -p /uploads

# Открываем порт
EXPOSE 8081

# Переменные окружения по умолчанию
ENV PORT=8081
ENV STORAGE_PATH=/uploads
ENV MIGRATIONS_DIR=/migrations

# Команда запуска
CMD ["./file-storage"]