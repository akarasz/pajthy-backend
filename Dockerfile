FROM golang:1.16-alpine AS builder

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download -x

COPY . .

RUN go build -o main ./cmd/standalone

FROM alpine:latest

COPY --from=builder /build/main .

EXPOSE 8000
CMD ["./main"]