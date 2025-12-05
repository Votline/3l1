FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git
ARG GITHUB_TOKEN
WORKDIR /build/order-service

RUN git config --global url."https://oauth2:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"

COPY order-service/go.mod order-service/go.sum ./
RUN export GOPRIVATE=github.com/Votline && go mod download -x

COPY order-service/ ./
RUN go build -o order-service .

FROM alpine:3.18
WORKDIR /order-service
COPY --from=builder /build/order-service/order-service .
CMD ["./order-service"]
