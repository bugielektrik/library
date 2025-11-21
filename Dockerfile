# Define the base image to use for building the application (Go 1.21.1 on Alpine Linux)
FROM golang:1.25-alpine as builder

# Install necessary dependencies, including OpenSSL and development tools
# RUN apk add --update --no-cache openssl curl g++ gcc libxslt-dev

# Fetch and save SSL certificate from proxy.golang.org to add to the system's trusted certificates
# RUN openssl s_client -showcerts -connect proxy.golang.org:443 -servername proxy.golang.org < /dev/null 2>/dev/null | openssl x509 -outform PEM > /usr/local/share/ca-certificates/ca.crt

# Set permissions for the saved SSL certificate and update the system's certificate authorities
# RUN chmod 644 /usr/local/share/ca-certificates/ca.crt && update-ca-certificates

# Set the working directory within the builder container
WORKDIR /build

# Copy the source code into the builder container
COPY . /build

# Build the Go application for Linux (amd64) with CGO enabled
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o library-service .

# Create a new stage for the final application image (based on Alpine Linux)
FROM alpine:3.22 as hoster

# Install necessary dependencies, including cURL and development tools
# RUN apk add --update --no-cache openssl curl g++ gcc libxslt-dev

# Copy configuration files, assets, templates, and the built application from the builder stage
COPY --from=builder /build/.env ./.env
COPY --from=builder /build/library-service ./library-service

# Define the entry point for the final application image
ENTRYPOINT [ "./library-service" ]
