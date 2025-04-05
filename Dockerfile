FROM golang:1.24-rc-bookworm AS go

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go run ./cmd/data-gen

RUN go build -ldflags='-s -w' -o ./tmp/bin/cluster ./cmd/cluster
RUN go build -ldflags='-s -w' -o ./tmp/bin/loadbalance ./cmd/loadbalance
RUN go build -ldflags='-s -w' -o ./tmp/bin/user ./cmd/user

FROM python:3.14-rc-bookworm

WORKDIR /app
COPY --from=go /app/tmp/bin/cluster ./tmp/bin/cluster
COPY --from=go /app/tmp/bin/loadbalance ./tmp/bin/loadbalance
COPY --from=go /app/tmp/bin/user ./tmp/bin/user
COPY --from=go /app/tmp/data ./tmp/data

# COPY requirements.txt .
# RUN pip install -r requirements.txt

COPY --from=go /app/runsim.sh ./runsim.sh
COPY gen_files.py .
CMD ["./runsim.sh"]
