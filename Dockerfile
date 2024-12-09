# Step 1: Use the official Go image to build the binary
FROM golang:1.23.1 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first (to cache dependencies)
COPY go.mod go.sum ./

# Ensure dependencies are downloaded and cached by running `go mod tidy`
RUN go mod tidy

# Copy the Go source code
COPY main.go ./

# Run `go get` explicitly to fetch dependencies before building
RUN go get github.com/eiannone/keyboard

# Build the Go binary (produces an executable named 'app')
RUN go build -o app .

# Step 2: Create a minimal image to run the binary
FROM alpine:latest

# Install any necessary dependencies (e.g., SSL certificates)
RUN apk --no-cache add ca-certificates

# Set the working directory in the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/app .

# Expose the port the app will run on (adjust if needed)
EXPOSE 8080

# Command to run the application
CMD ["./app"]
