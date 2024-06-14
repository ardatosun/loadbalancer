# Use official golang image as the base image
FROM golang:1.22.4 AS builder

# Set the current working directory inside the container
WORKDIR /app

COPY go.mod .
RUN go mod download
# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o main .

# Start a new stage from scratch
FROM alpine:latest

# Set the working directory to /app in the container
WORKDIR /app

# Copy the compiled Go executable from the builder stage
COPY --from=builder /app/main .

# Expose port 80 on which the HTTP server will run
EXPOSE 80

# Command to run the HTTP server
CMD ["./main"]
