FROM golang:alpine AS builder

RUN apk add --no-cache build-base tzdata

WORKDIR /app

COPY . .

RUN go build -ldflags="-w -s" .

FROM alpine:3

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/rinha-backend-2024 /usr/bin/

ENTRYPOINT ["rinha-backend-2024"]
