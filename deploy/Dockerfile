FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/server /bin/server
COPY configs/ /app/configs/

WORKDIR /app

EXPOSE 8080 50051

ENTRYPOINT ["/bin/server"]
CMD ["-config", "/app/configs/config.prod.yaml"]
