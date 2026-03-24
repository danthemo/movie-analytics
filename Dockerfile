FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /movie-analytics ./cmd/server

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /movie-analytics /app/movie-analytics
COPY --from=builder /app/frontend /app/frontend

EXPOSE 8080

CMD ["/app/movie-analytics"]
