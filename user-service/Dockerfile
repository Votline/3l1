FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git
ARG GITHUB_TOKEN
WORKDIR /build/user-service

RUN git config --global url."https://oauth2:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"


COPY user-service/go.mod user-service/go.sum ./
RUN export GOPRIVATE=github.com/Votline && go mod download -x

COPY user-service/ ./
RUN go build -o user-service .

FROM alpine:3.18
WORKDIR /user-service
COPY --from=builder /build/user-service/user-service .
CMD ["./user-service"]
