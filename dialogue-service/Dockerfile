# syntax=docker/dockerfile:1

FROM golang:1.18-bullseye

COPY internal dialogue-service/internal
COPY go.mod dialogue-service/go.mod
COPY main.go dialogue-service/main.go
COPY config dialogue-service/config

WORKDIR ./dialogue-service

RUN go mod download
RUN go build -mod=mod -o ./dialogue-service

EXPOSE 80

ENV CONFIG_FILE=./config/dev.yaml

CMD [ "./dialogue-service" ]
