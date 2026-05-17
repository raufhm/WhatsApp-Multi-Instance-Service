FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o whatsapp-api .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/whatsapp-api .
COPY migrations /app/migrations
EXPOSE 8080
CMD ["./whatsapp-api"]
