FROM golang:1.18-bullseye

ENV APP_HOME /go/src/linkcrawler
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"
EXPOSE 8010
CMD ["go", "run", "cmd/linkcrawler-server/main.go"]