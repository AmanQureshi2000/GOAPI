# --- Stage 1: Build the binary ---
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
# (Doing this before copying source code allows Docker to cache layers)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app as a static binary named "server"
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# --- Stage 2: Final lightweight image ---
FROM alpine:latest

# Install CA certificates (required for connecting to external DBs like Aiven via SSL/TLS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .

# Render provides the PORT env variable; we just need to make sure the app uses it
EXPOSE 8080

# Run the binary
CMD ["./server"]