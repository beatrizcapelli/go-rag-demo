# ===== Stage 1: Build =====
FROM golang:1.25 AS builder

WORKDIR /app

# Copy go mod files first and download deps (better cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# ===== Stage 2: Run =====
FROM alpine:3.20

WORKDIR /app

# Copy binary and frontend assets from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/frontend /app/frontend

# Port your Go app listens on
EXPOSE 8080

# Run the binary
CMD ["/app/server"]
