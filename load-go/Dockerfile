FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod .
COPY main.go ./
COPY static ./static
RUN go build -o /load-generator

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /load-generator /app/load-generator
COPY static ./static
EXPOSE 8080
CMD ["/app/load-generator"]