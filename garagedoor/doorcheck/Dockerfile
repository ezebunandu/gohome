FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o doorcheck .


FROM gcr.io/distroless/static:nonroot
WORKDIR /app
USER nonroot:nonroot
COPY --chown=nonroot:nonroot --from=builder /app/doorcheck .
ENTRYPOINT ["/app/doorcheck"]
