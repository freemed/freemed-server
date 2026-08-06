# Multi-stage build for FreeMED Go backend
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum go.work ./
COPY api/go.mod api/go.sum ./api/
COPY billing/go.mod billing/go.sum ./billing/
COPY cmd/freemed-server/go.mod cmd/freemed-server/go.sum ./cmd/freemed-server/
COPY common/go.mod common/go.sum ./common/
COPY config/go.mod config/go.sum ./config/
COPY model/go.mod model/go.sum ./model/

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /freemed ./cmd/freemed-server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /freemed /freemed
COPY config.yml /config.yml

EXPOSE 3000
ENTRYPOINT ["/freemed", "-config", "/config.yml"]
