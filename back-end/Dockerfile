FROM golang as builder
WORKDIR /build
COPY *go /build

RUN CGO_ENABLED=0 go build -o backend-server *.go

FROM alpine
WORKDIR /app
COPY --from=builder /build/backend-server /app

CMD ["./backend-server"]
