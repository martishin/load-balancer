# Build stage
FROM golang:1.21.4-alpine3.18 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . ./
RUN GOOS=linux go build -o load-balancer ./cmd

# Run stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/load-balancer /app/load-balancer
EXPOSE 80
CMD ["/app/load-balancer"]
