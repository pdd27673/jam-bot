# Use the official Golang image as the build stage
FROM golang:1.22-alpine AS builder


# Set the Current Working Directory inside the container
WORKDIR /app

# Install git and other dependencies
RUN apk update && apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
WORKDIR /app/cmd/bot
RUN go build -o /discord-bot .

# Start a new stage from scratch
FROM alpine:latest

# Install necessary CA certificates
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /discord-bot .

# Copy documentation files
COPY --from=builder /app/README.md .

# Copy the config file
COPY --from=builder /app/internal/config/config.yml ./internal/config/

# Expose port (use the same port as in your code)
EXPOSE 8080

# Command to run the executable
CMD ["./discord-bot"]