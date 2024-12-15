# use official Golang image as the base
FROM golang:1.22-alpine

# # Make a directory for the app
# RUN mkdir /app

# # Copy the go files into the app directory
# COPY ./**/*.go /app

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN ls -la .

# Build the Go app
RUN go build -o discord-bot .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./discord-bot"]