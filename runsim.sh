#!/usr/bin/bash

echo "Removing existing output"
rm -fr ./tmp/output

echo "Creating log dir"
mkdir -p log/lb/
mkdir -p log/cluster/
mkdir -p log/user/

echo "Compiling binaries"
go build -ldflags='-s -w' -o tmp/cluster ./cmd/cluster
go build -ldflags='-s -w' -o tmp/loadbalance ./cmd/loadbalance
go build -ldflags='-s -w' -o tmp/user ./cmd/user

runsim() {
	lblogfile="log/lb/lb-$(basename "$1" .yaml).log"
	clusterlogfile="log/cluster/cluster-$(basename "$1" .yaml).log"
	userlogfile="log/user/user-$(basename "$1" .yaml).log"

	# kill the process binding port 8000 (if exists)
	pid=$(lsof -t -i:8000)
	[[ -z $pid ]] || kill -9 $(lsof -t -i:8000)

	export CONFIG_PATH=$1
	sleep 3
	./tmp/loadbalance &
	sleep 1
	./tmp/cluster &
	sleep 1
	./tmp/user
}

for file in ./config/*; do
	if [ -f "$file" ]; then
		echo "Running simulation: $file"
		runsim $file
	fi
done
