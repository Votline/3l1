FROM golang:1.25.5-alpine AS builder

RUN apk add --no-cache git
ARG GITHUB_TOKEN
WORKDIR /build/user-service

COPY user-service/go.mod user-service/go.sum ./
RUN go mod download -x

COPY user-service/ ./
RUN go build -o user-service .

FROM alpine:3.18
WORKDIR /user-service
COPY --from=builder /build/user-service/user-service .
CMD ["./user-service"]
