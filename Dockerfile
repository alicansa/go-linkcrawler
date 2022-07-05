FROM golang:1.18-bullseye

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

ENV APP_HOME /go/src/linkcrawler
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"
EXPOSE 8010
CMD ["bee", "run"]