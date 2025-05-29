FROM golang:1.24 AS builder

WORKDIR /TaskFlow

COPY . .

WORKDIR /TaskFlow/task-api

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/task-api
RUN chmod +x main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /TaskFlow/task-api/main .
COPY deployments/.env .
COPY task-api/wait-for-rabbitmq.sh .

RUN apk add --no-cache busybox-extras
RUN chmod +x wait-for-rabbitmq.sh

EXPOSE 8080

CMD ["./wait-for-rabbitmq.sh", "./main"]