FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /analytics ./cmd/server

FROM gcr.io/distroless/static-debian12

COPY --from=builder /analytics /analytics

EXPOSE 8080

ENTRYPOINT ["/analytics"]
