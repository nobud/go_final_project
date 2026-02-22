FROM golang:1.24.4 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler .

FROM alpine:latest



RUN mkdir -p /app

COPY --from=builder /build/scheduler /app/scheduler
COPY ./web /app/web

# Порт можно изменить через переменную окружения TODO_PORT при запуске
ENV TODO_PORT=7540

# порт по умолчанию
EXPOSE 7540

WORKDIR /app
CMD ["./scheduler"]






