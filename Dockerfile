# Multi-stage build for FreeMED Go backend
# The builder image must match the go directive in go.mod: the stdlib advisories
# that govulncheck reports against go1.26.0 are fixed in go1.26.6, so pinning a
# floating 1.26 tag would reintroduce them whenever the patch lags.
FROM golang:1.26.7-alpine AS builder

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

FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata && apk upgrade --no-cache
COPY --from=builder /freemed /freemed
COPY config.yml /config.yml

EXPOSE 3000
ENTRYPOINT ["/freemed", "-config", "/config.yml"]
