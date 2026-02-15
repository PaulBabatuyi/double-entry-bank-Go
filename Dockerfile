#  Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ledger cmd/main.go

#  Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /ledger .
COPY --from=builder /app/docs ./docs
COPY .env.example .env  
# copy template; override in production

EXPOSE 8080
CMD ["/app/ledger"]