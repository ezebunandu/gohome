FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o lightscheduler lightScheduler.go config.go

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D lightscheduler
WORKDIR /app
COPY --chown=lightscheduler --from=builder /app/lightscheduler .

RUN chmod +x /app/lightscheduler

EXPOSE 8100
ENTRYPOINT [ "/app/lightscheduler" ]
