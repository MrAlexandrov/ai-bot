FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
ENV GONOSUMDB=*
ENV GOFLAGS=-mod=mod
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bot .

CMD ["./bot"]
