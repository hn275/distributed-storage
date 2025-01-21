FROM golang:1.24-rc-bookworm
WORKDIR /node
COPY . .
RUN go build ./cmd/discovery
CMD ["./discovery"]
