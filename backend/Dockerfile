FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY logx/go.mod logx/
COPY backend/go.mod backend/

ENV GOFLAGS=-mod=mod
RUN cd backend && go mod download

COPY logx/ logx/
COPY backend/ backend/

RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/backend .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /out/backend .

EXPOSE 8080
CMD ["./backend"]
