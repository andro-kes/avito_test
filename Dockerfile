FROM golang:1.24.2-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]