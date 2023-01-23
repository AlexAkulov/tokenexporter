FROM golang:1.18.0 as builder

COPY . /app

WORKDIR /app

RUN GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0 \
    go build -ldflags '-w -extldflags "-static"' -o tokenexporter


FROM alpine:3.15

RUN addgroup -S -g 101 app \
    && adduser -S -h /app -s /bin/bash -G app -g "app" -u 101 app

RUN apk update && apk add --no-cache ca-certificates tzdata bash curl && update-ca-certificates

COPY --from=builder /app/tokenexporter /app/tokenexporter

USER app
ENV LISTEN ":9015"
ENV CONFIG_FILE "/etc/tokenexporter/config.yml"

CMD ["/app/tokenexporter"]
