FROM golang:1.26.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o subscriber .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/subscriber .
EXPOSE 6002
CMD ["./subscriber"]