# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o lychee ./cmd/runner

# ---- Runtime Stage ----
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /build/lychee .

EXPOSE 11434

CMD ["./lychee", "serve"]
