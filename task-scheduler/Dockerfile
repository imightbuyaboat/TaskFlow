FROM golang:1.24 AS builder

WORKDIR /TaskFlow

COPY . .

WORKDIR /TaskFlow/task-scheduler

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/task-scheduler
RUN chmod +x main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /TaskFlow/task-scheduler/main .
COPY deployments/.env .
COPY deployments/wait-for-services.sh .
COPY task-scheduler/config.json .

RUN apk add --no-cache busybox-extras
RUN chmod +x wait-for-services.sh

EXPOSE 8080

CMD ["./wait-for-services.sh", "./main"]