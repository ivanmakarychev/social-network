# syntax=docker/dockerfile:1

FROM golang:1.18-bullseye

# RUN git clone https://github.com/ivanmakarychev/social-network.git

COPY internal social-network/internal
COPY go.mod social-network/go.mod
COPY main.go social-network/main.go
COPY config social-network/config

WORKDIR ./social-network

RUN go mod download
RUN go build -mod=mod -o ./social-network

EXPOSE 80

ENV CONFIG_FILE=./config/dev.yaml

CMD [ "./social-network" ]
