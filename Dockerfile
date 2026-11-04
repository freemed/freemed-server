# Multi-stage build for FreeMED Go backend
# The builder image must equal the go directive in go.mod. Official golang
# images set GOTOOLCHAIN=local, so a builder patch older than go.mod's go
# directive fails outright ("go.mod requires go >= X") instead of upgrading.
FROM golang:1.26.10-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
# One COPY per module in go.work, go.sum included so the download step verifies
# hashes in the container. billing/ is deliberately absent: it is not a go.work
# module, nothing imports it, and it cannot resolve (its go.mod has no replace
# for the root module), so its go.sum can never be generated.
COPY go.mod go.sum go.work ./
COPY api/go.mod api/go.sum ./api/
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
