# Use the official Golang image as the build stage
FROM golang:1.22-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Install git for fetching dependencies
RUN apk update && apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

# Change directory to where main.go is located
WORKDIR /app/cmd/bot

# Build the Go app
RUN go build -o /discord-bot .

# Start a new stage from scratch
FROM alpine:latest

# Install necessary CA certificates
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /discord-bot .

# Expose port 8080 to the outside world (Render uses the PORT environment variable)
EXPOSE 8080

# Command to run the executable
CMD ["./discord-bot"]