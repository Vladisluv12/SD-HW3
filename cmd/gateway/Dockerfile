FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем go модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/gateway

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/gateway .

EXPOSE 8080
ENV PORT=8080

CMD ["./gateway"]