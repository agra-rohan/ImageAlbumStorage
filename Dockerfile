FROM golang:alpine

WORKDIR /build

COPY go.mod .
COPY go.mod .

RUN go mod download

COPY . .

# Build the application
RUN go build -o main .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /app

# Copy binary from build to main folder
RUN cp /build/main .

# Command to run when starting the container
CMD ["/app/main"]
