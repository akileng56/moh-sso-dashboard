FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o sso-server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/sso-server .
EXPOSE 8080
CMD ["./sso-server"]
