FROM golang:1.24-rc-bookworm
WORKDIR /node
COPY . .
RUN go build ./cmd/node
RUN apt-get update && apt-get install -y iproute2
COPY ./dockerfiles/start_node.sh /start_node.sh
RUN chmod +x /start_node.sh
CMD ["/start_node.sh"]
