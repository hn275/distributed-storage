lb:
	go run ./cmd/loadbalance/

cluster:
	go run ./cmd/cluster/

user:
	go run ./cmd/user/

compile:
	go build -ldflags='-s -w' -o ./tmp/bin/cluster ./cmd/cluster
	go build -ldflags='-s -w' -o ./tmp/bin/loadbalance ./cmd/loadbalance
	go build -ldflags='-s -w' -o ./tmp/bin/user ./cmd/user
