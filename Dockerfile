FROM golang:1.27-alpine3.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-server ./cmd/api-server

FROM alpine:3.24
RUN adduser -S -D -H app
USER app
COPY --from=build /out/api-server /app/api-server
EXPOSE 8090
ENTRYPOINT ["/app/api-server"]