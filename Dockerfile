# builder image
FROM golang:1.20.4-alpine as builder
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# generate clean, final image for end users
FROM alpine:3.18.0 as hoster
COPY --from=builder /build/app ./app
COPY --from=builder /build/.env* ./.env
COPY --from=builder /build/migrations/ ./migrations/

# executable
ENTRYPOINT [ "./app" ]
