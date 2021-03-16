FROM golang:1.15 AS builder

WORKDIR /build
COPY . .
RUN go get
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o app

FROM alpine
WORKDIR /root/
COPY --from=builder /build/app .

EXPOSE 60001
CMD ["./app"]