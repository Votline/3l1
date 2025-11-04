FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git
ARG GITHUB_TOKEN
WORKDIR /build/api-gateway

RUN git config --global url."https://oauth2:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"

COPY api-gateway/go.mod api-gateway/go.sum ./
RUN export GOPRIVATE=github.com/Votline && go mod download -x

COPY api-gateway/ ./
RUN go build -o api-gateway

FROM alpine:3.18
WORKDIR /api-gateway
COPY --from=builder /build/api-gateway/api-gateway .
CMD ["./api-gateway"]
