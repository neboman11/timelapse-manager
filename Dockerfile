# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY api ./api
COPY docs ./docs
COPY models ./models
RUN go build

EXPOSE 3001

CMD [ "/app/timelapse-manager" ]
