# Build stage
FROM golang:1.16.3-alpine3.13 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the current directory contents into the container
COPY . .

# Install dependencies and build the Go application
RUN go get -d -v ./...
RUN go build -o main .

# Final stage - minimal image with just the built binary
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your app will run on (in this case 8003)
EXPOSE 8003

# Start the application
CMD ["./main"]
