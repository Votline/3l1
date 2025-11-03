FROM golang:1.24.5-alpine

WORKDIR /api-gateway
COPY api-gateway/ ./
RUN go build -o api-gateway .

CMD ["./api-gateway"]
