FROM golang:1.27.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build  -o /app/auth-service ./cmd/auth

FROM alpine:3.22 AS runtime

WORKDIR /app

COPY --from=builder /app/auth-service .

RUN adduser -D -H appuser

USER appuser

EXPOSE 8080

ENTRYPOINT [ "./auth-service" ]