FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build ./cmd/client/

FROM scratch
COPY --from=builder /app/client /client
ENTRYPOINT ["/client"]
