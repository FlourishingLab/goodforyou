# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
# This leverages Docker's layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .


# Build the Go application, creating a statically linked binary
# This is important for running in a minimal 'distroless' image
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /server .

# Stage 2: Create the final, minimal image
FROM gcr.io/distroless/static-debian12

# Copy the compiled binary from the builder stage
COPY --from=builder /server /server
COPY questions/questions.csv /questions/questions.csv

# Google Cloud Run sets the PORT environment variable, which defaults to 8080.
# Your application should listen on this port.
EXPOSE 8080

# Set the entrypoint for the container
CMD ["/server"]