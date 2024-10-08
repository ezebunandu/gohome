FROM docker.io/golang:1.23.2 AS BUILDER
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w"

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D picotempexport
WORKDIR /app
COPY --chown=picotempexport --from=builder  /app/picotempexport .
EXPOSE 3030
CMD ["/app/picotempexport"]