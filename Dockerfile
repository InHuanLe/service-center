# Multi-stage build for service-center registry server

FROM golang:1.22 AS builder
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/service-center ./...

# Final minimal image
FROM gcr.io/distroless/base-debian12
COPY --from=builder /bin/service-center /service-center

EXPOSE 50051
ENTRYPOINT ["/service-center"]
