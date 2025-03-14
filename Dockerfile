# syntax=docker/dockerfile:1

FROM golang:latest AS build

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
COPY --from=build build ./

VOLUME /data

EXPOSE 3001

ENTRYPOINT [ "/app/timelapse-manager" ]
