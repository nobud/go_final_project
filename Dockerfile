FROM golang:1.25 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o scheduler .

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /build/scheduler ./scheduler
COPY web/ ./web/

ENV TODO_PORT=7540
EXPOSE 7540

CMD ["./scheduler"]






