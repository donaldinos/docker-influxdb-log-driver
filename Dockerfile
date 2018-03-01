FROM golang:1.9-alpine

RUN apk update && \
	apk add git

WORKDIR /go/src/docker-influxdb-log-driver

RUN go get github.com/docker/docker/api/types/plugins/logdriver
RUN go get github.com/docker/docker/daemon/logger
RUN go get github.com/docker/docker/daemon/logger/loggerutils
RUN go get github.com/docker/go-plugins-helpers/sdk
RUN go get github.com/pkg/errors
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/tonistiigi/fifo
RUN go get github.com/gogo/protobuf/io
RUN go get github.com/influxdata/influxdb

COPY . /go/src/docker-influxdb-log-driver
RUN go get
RUN go build --ldflags '-extldflags "-static"' -o /usr/bin/docker-influxdb-log-driver

#-- Second container --
FROM alpine:latest

RUN apk --no-cache add --update tzdata

ENV TZ=Europa/Prague

WORKDIR /usr/bin

COPY --from=0 /usr/bin/docker-influxdb-log-driver .
