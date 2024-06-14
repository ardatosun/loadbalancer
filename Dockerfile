# Start with a base image that includes Go
FROM golang:1.22.4 AS builder

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy and download dependencies using go mod
COPY go.mod .
RUN go mod download

# Copy the entire source code
COPY . .

# Build the Go application
# RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

# Start a new stage from scratch
FROM alpine:latest

# Copy the built executable from the previous stage
# COPY --from=builder /app/loadbalancer /loadbalancer

# Set the working directory inside the container
WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/loadbalancer .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./loadbalancer"]

