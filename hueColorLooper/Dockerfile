FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o colorlooper colorlooper.go

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D colorlooper
WORKDIR /app
COPY --chown=colorlooper --from=builder /app/colorlooper .

# set the timezone
ENV TZ=America/Edmonton

RUN apk add --no-cache tzdata && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone

RUN chmod +x /app/colorlooper

EXPOSE 3005
ENTRYPOINT [ "/app/colorlooper" ]
