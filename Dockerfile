FROM golang:1.25-alpine AS builder

COPY . /app

WORKDIR /app

RUN go build

FROM scratch

COPY --from=builder /app/backend .

CMD ["./backend"]