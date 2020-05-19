FROM golang:1.13

WORKDIR /go/src/bast
COPY . .

RUN go build
RUN mkfifo pin-pipe
RUN mkfifo card-pipe

CMD ["./lock-firmware"]
