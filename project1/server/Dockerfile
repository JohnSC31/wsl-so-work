FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .

# Instala curl para healthchecks
RUN apk add --no-cache curl

EXPOSE 8080
CMD ["./server"]