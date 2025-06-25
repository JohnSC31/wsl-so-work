FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o dispatcher .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dispatcher .

# Instalar Docker CLI para poder crear contenedores
RUN apk add --no-cache docker-cli

EXPOSE 8080
CMD ["./dispatcher", "--init-workers=3"]