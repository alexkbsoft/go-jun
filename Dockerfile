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

# Встраиваем CA-сертификаты MAX
COPY certs/Russian_Trusted_Root_CA.cer /usr/local/share/ca-certificates/
COPY certs/Russian_Trusted_Sub_CA.cer /usr/local/share/ca-certificates/
COPY certs/Russian_Trusted_Sub_CA_2024.cer /usr/local/share/ca-certificates/
RUN cat /usr/local/share/ca-certificates/Russian_Trusted_Root_CA.cer \
    /usr/local/share/ca-certificates/Russian_Trusted_Sub_CA.cer \
    /usr/local/share/ca-certificates/Russian_Trusted_Sub_CA_2024.cer \
    >> /etc/ssl/certs/ca-certificates.crt

EXPOSE 8090
