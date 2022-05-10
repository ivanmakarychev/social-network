# syntax=docker/dockerfile:1

FROM golang:1.18-bullseye

RUN git clone https://github.com/ivanmakarychev/social-network.git

WORKDIR ./social-network

RUN go mod download
RUN go build -mod=mod -o ./social-network

EXPOSE 80

CMD [ "./social-network" ]
