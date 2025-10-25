FROM golang:latest AS builder

WORKDIR /app

ARG goproxy_url
ENV GOPROXY=http://$goproxy_url,https://proxy.golang.org
ENV GONOSUMDB=apps-hosting.com
ENV GOINSECURE=http://$goproxy_url

COPY go.mod go.sum .

RUN go mod download

COPY . .
RUN go build -o main .


FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates git \
 && update-ca-certificates \
 && rm -rf /var/lib/apt/lists/*
 
COPY --from=builder /app/main /app/main

EXPOSE 8080
CMD ["/app/main"]
