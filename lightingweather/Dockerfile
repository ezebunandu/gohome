FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o lightweather lightingweather.go config.go

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D lightweather
WORKDIR /app
COPY --chown=lightweather --from=builder /app/lightweather .
EXPOSE 3040
ENTRYPOINT [ "/app/lightweather" ]