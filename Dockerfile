FROM golang:1.21-alpine AS builder

# Install Templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Templ files and build
RUN templ generate && go build -o bin/server ./cmd/server

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary and static assets
COPY --from=builder /app/bin/server .
COPY --from=builder /app/web/static ./web/static

# Expose port
EXPOSE 8080

CMD ["./server"]