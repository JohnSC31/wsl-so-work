# Etapa de construcción
FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dispatcher .

# Etapa final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dispatcher .
RUN ls -la && chmod +x dispatcher

CMD ["./dispatcher"]