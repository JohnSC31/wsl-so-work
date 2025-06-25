FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .

# Instalar curl para healthchecks (opcional)
RUN apk add --no-cache curl

EXPOSE 8080

# Variables de entorno para auto-registro
ENV PORT=8080
ENV DISPATCHER_URL=http://dispatcher:8080
# HOSTNAME se inyecta automáticamente por Docker

CMD ["./server"]