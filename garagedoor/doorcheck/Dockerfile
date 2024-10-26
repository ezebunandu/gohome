FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS-linux GOARCH=arm go build -ldflags="-s -w" -o doorcheck .

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D doorcheck
WORKDIR /app
COPY --chown=doorcheck --from=builder /app/doorcheck .
ENTRYPOINT [ "/app/doorcheck" ]
