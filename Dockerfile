FROM golang:1.24.4 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler .

FROM ubuntu:latest

RUN apt-get update && apt-get install -y sqlite3

RUN mkdir -p /app

COPY --from=builder /build/scheduler /app/scheduler
COPY ./web /app/web

WORKDIR /app
EXPOSE 7540
CMD ["./scheduler"]






