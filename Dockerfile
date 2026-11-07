FROM golang:1.27-alpine3.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/outbox-poller ./cmd/outbox-poller
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/inbox-poller ./cmd/inbox-poller


FROM alpine:3.24
RUN adduser -S -D -H app

COPY --from=build /out/api-server /app/api-server
COPY --from=build /out/outbox-poller /app/outbox-poller
COPY --from=build /out/inbox-poller /app/inbox-poller

# Скачиваем CA-сертификаты Russian Trusted CA с CDN Max
RUN curl -sL "http://nuc-cdp.voskhod.ru/cdp/subca_ssl_rsa2024.crt" -o /tmp/max-subca.crt && \
    curl -sL "http://nuc-cdp.voskhod.ru/cdp/rootca_ssl_rsa2022.crt" -o /tmp/max-rootca.crt && \
    cat /tmp/max-subca.crt /tmp/max-rootca.crt >> /etc/ssl/certs/ca-certificates.crt

EXPOSE 8090
