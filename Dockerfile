FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dnstt-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dnstt-client ./cmd/client

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/dnstt-server /usr/local/bin/
COPY --from=builder /build/dnstt-client /usr/local/bin/

EXPOSE 53/udp

# Default to server mode
ENTRYPOINT ["dnstt-server"]
CMD ["-l", "0.0.0.0:53", "-h", "16", "-v"]
