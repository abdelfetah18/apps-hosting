FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum .

RUN go mod download

COPY . .
RUN go build -o go-registry .

FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates git \
 && update-ca-certificates \
 && rm -rf /var/lib/apt/lists/*
 
WORKDIR /app

COPY --from=builder /app/go-registry /app/go-registry

EXPOSE 8080
CMD ["/app/go-registry","server"]