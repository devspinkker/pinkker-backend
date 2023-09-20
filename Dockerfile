# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the working directory
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the entire project to the working directory
COPY . .

# Copy the .env file from the current directory to the working directory in the container
COPY .env .

# Build the Go application (assuming main.go is inside the cmd directory)
RUN go build -o myapp ./cmd

# Expose a port (if your application listens on a specific port)
EXPOSE 8080

# Command to run the application
CMD ["./myapp"]