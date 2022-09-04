# syntax=docker/dockerfile:1

FROM golang:1.18-bullseye

COPY internal counter-service/internal
COPY go.mod counter-service/go.mod
COPY main.go counter-service/main.go
COPY config counter-service/config

WORKDIR ./counter-service

RUN go mod download
RUN go build -mod=mod -o ./counter-service

EXPOSE 80

ENV CONFIG_FILE=./config/dev.yaml

CMD [ "./counter-service" ]
