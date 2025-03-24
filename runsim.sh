#!/usr/bin/bash

mkdir -p log/lb/
mkdir -p log/cluster/
mkdir -p log/user/

runsim() {
	lblogfile="log/lb/lb-$(basename "$1" .yaml).log"
	clusterlogfile="log/cluster/cluster-$(basename "$1" .yaml).log"
	userlogfile="log/user/user-$(basename "$1" .yaml).log"

	export CONFIG_PATH=$1
	nohup go run ./cmd/loadbalance >$lblogfile 2>&1 &
	nohup go run ./cmd/cluster >$clusterlogfile 2>&1 &
	go run ./cmd/user >$userlogfile 2>&1
}

for file in ./config/*; do
	if [ -f "$file" ]; then
		echo "Running simulation: $file"
		runsim $file
		sleep 2 # sleep so the load balancer and cluster can shutdown properly
	fi
done
