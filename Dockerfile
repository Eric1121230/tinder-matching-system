FROM golang:1.22-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go test ./... && \
    go build -ldflags="-w -s" -o tinder-server ./cmd/server/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/tinder-server .

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["./tinder-server"]
