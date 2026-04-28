FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY *.go .
RUN go build -o event-processor .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/event-processor .
EXPOSE 5002
CMD ["./event-processor"]
