FROM golang:1.24-rc-bookworm

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go run ./cmd/data-gen

RUN go build -ldflags='-s -w' -o ./tmp/bin/cluster ./cmd/cluster
RUN go build -ldflags='-s -w' -o ./tmp/bin/loadbalance ./cmd/loadbalance
RUN go build -ldflags='-s -w' -o ./tmp/bin/user ./cmd/user

CMD ["./runsim.sh"]
