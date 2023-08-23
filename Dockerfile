# builder image
FROM golang:1.18-alpine as builder
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOOS=linux go build -o library-service .

# generate clean, final image for end users
FROM alpine:3.18.0 as hoster
COPY --from=builder /build/.env* ./.env
COPY --from=builder /build/migrations/ ./migrations/
COPY --from=builder /build/library-service ./library-service

# executable
ENTRYPOINT [ "./library-service" ]
