FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git
ARG GITHUB_TOKEN
WORKDIR /build/api-gateway

COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download -x

COPY api-gateway/ ./
RUN go build -o api-gateway

FROM alpine:3.18
WORKDIR /api-gateway
RUN mkdir -p logs
COPY --from=builder /build/api-gateway/api-gateway .
CMD ["./api-gateway"]
