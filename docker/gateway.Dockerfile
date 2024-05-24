FROM golang:1.22-alpine as builder

LABEL stage=gobuilder

ENV CGO_ENABLED = 0
ENV GOOS linux

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

COPY . .

RUN go mod download && go mod verify

RUN go build -ldflags="-s -w" -o /app/gateway backend/cmd/gateway/main.go

FROM alpine

RUN apk update && apk upgrade

RUN rm -rf /var/cache/apk/* && \
    rm -rf /tmp/*

RUN adduser -D appuser
USER appuser

WORKDIR /app

COPY --from=builder /app/gateway /app/gateway
COPY --from=builder /build/.env /app

ENTRYPOINT ["/app/gateway"]

