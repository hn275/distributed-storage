#!/usr/bin/bash

log_dir=tmp/log
bin_dir=tmp/bin

runsim() {
	log_lb="${log_dir}/lb/lb-$(basename "$1" .yml).log"
	log_cluster="${log_dir}/cluster/cluster-$(basename "$1" .yml).log"
	log_user="${log_dir}/user/user-$(basename "$1" .yml).log"

	export CONFIG_PATH=$1

	echo "Running simulation: $1"

	echo "Starting lb"
	${bin_dir}/loadbalance >$log_lb 2>&1 &
	pid_lb=$!
	sleep 1

	echo "Starting cluster"
	${bin_dir}/cluster >$log_cluster 2>&1 &
	pid_cluster=$!
	sleep 1

	echo "Starting user"
	${bin_dir}/user |& tee $log_user

	echo "Waiting for lb shutdown"
	wait $pid_lb

	echo "Waiting for cluster shutdown"
	wait $pid_cluster
}

echo "Removing existing output"
rm -fr ./tmp/output

echo "Creating bin dir"
mkdir -p ${bin_dir}

echo "Creating log dir"
mkdir -p ${log_dir}/lb/
mkdir -p ${log_dir}/cluster/
mkdir -p ${log_dir}/user/

echo "Compiling binaries"
go build -ldflags='-s -w' -o ${bin_dir}/cluster ./cmd/cluster
go build -ldflags='-s -w' -o ${bin_dir}/loadbalance ./cmd/loadbalance
go build -ldflags='-s -w' -o ${bin_dir}/user ./cmd/user

file=$1

if [[ -z $file ]]; then
	echo "Running all files in ./config/"

	for file in ./config/*; do
		if [ -f "$file" ]; then
			runsim $file
		fi
	done
else
	runsim $file
fi

echo "Generating plot: user"
./gen_files.py user
echo "Generating plot: cluster"
./gen_files.py cluster
echo "Generating plot: lb"
./gen_files.py lb
