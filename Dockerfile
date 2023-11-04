# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS build

WORKDIR /build
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY api ./api
COPY docs ./docs
COPY models ./models
RUN go build

FROM linuxserver/ffmpeg:version-6.0-cli
WORKDIR /app
COPY --from=build app ./

VOLUME /data

EXPOSE 3001

CMD [ "/app/timelapse-manager" ]
