# Build stage
FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o exporter-agent cmd/exporter-agent/main.go

# Final stage using Alpine
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/exporter-agent /app/exporter-agent
EXPOSE 8080
ENTRYPOINT ["/app/exporter-agent"]
