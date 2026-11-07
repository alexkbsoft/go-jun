FROM golang:1.27-alpine3.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/outbox-poller ./cmd/outbox-poller
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/inbox-poller ./cmd/inbox-poller


FROM alpine:3.24

# 1. Устанавливаем пакет для обновления сертификатов
RUN apk --no-cache add ca-certificates && update-ca-certificates

RUN adduser -S -D -H app

COPY --from=build /out/api-server /app/api-server
COPY --from=build /out/outbox-poller /app/outbox-poller
COPY --from=build /out/inbox-poller /app/inbox-poller

# 2. Копируем CA-сертификаты Russian Trusted CA от Госуслуг/Max
COPY certs/*.cer /usr/local/share/ca-certificates/

# 3. Регистрируем добавленные сертификаты в системе
RUN update-ca-certificates

EXPOSE 8090
