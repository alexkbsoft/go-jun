FROM golang:1.27-alpine3.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/outbox-poller ./cmd/outbox-poller


FROM alpine:3.24
RUN adduser -S -D -H app
USER app
COPY --from=build /out/api-server /app/api-server
COPY --from=build /out/outbox-poller /app/outbox-poller
EXPOSE 8090