FROM golang:1.9 as builder
WORKDIR /go/src/
COPY server.go .
RUN GOOS=linux CGO_ENABLED=0 go build -o server .

FROM alpine:latest
WORKDIR /root/
EXPOSE 8080
COPY --from=builder /go/src/server .
COPY words .
ENTRYPOINT ["./server"]
