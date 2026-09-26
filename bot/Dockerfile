FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY logx/go.mod logx/
COPY bot/go.mod bot/
COPY bot/go.sum bot/

ENV GOFLAGS=-mod=mod
RUN cd bot && go mod download

COPY logx/ logx/
COPY bot/ bot/

RUN cd bot && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/bot .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /out/bot .

CMD ["./bot"]
