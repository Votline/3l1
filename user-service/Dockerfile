FROM golang:1.24.5-alpine

WORKDIR /user-service
COPY user-service/ ./
RUN go build -o user-service .

CMD ["./user-service"]
