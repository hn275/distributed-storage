#!/usr/bin/python3

import re 
import sys
import matplotlib.pyplot as plt
import os
import random

ALG_OPTIONS = (("rr", "simple-round-robin"), ("lc", "least-connections"), ("lrt", "least-response-time"))
HOMOG_OPTIONS = (True, False)
INTERVAL = 10
FILE_SZ = ("s", "m", "l")
RATES = (10, 100, 1000)
LATENCY_OPTIONS = (0, 100)
USER_DIR = "tmp/output/user"

def write_random_numbers(filename="output.txt"):
    with open(filename, "w") as file:
        for _ in range(100):
            num = random.randint(0, 1000)
            file.write(f"{num},256\n")

class ClientExp:
    def __init__(self, alg, net_delay=None, homog=None, interval=None, fsz=None, fsz_bytes=None, rate=None, avg_serv_time=None):
        self.alg = alg
        self.net_delay = net_delay
        self.homog = homog
        self.interval = interval
        self.fsz = fsz
        self.fsz_bytes = fsz_bytes
        self.rate = rate
        self.avg_serv_time = avg_serv_time
    
    def __repr__(self):
        return f"ClientExp(alg={self.alg}, net_delay={self.net_delay}, homog={self.homog}, interval={self.interval}, fsz={self.fsz}, rate={self.rate}, avg_serv_time={self.avg_serv_time}"

def gen_config(name, algo, homog, latency, interval, files):
    config = f"user:\n" + \
             f"  small: {files[0]}\n" + \
             f"  medium: {files[1]}\n"+ \
             f"  large: {files[2]}\n" + \
             f"  x-large: 0\n" + \
             f"  xx-large: 0\n" + \
             f"  interval: {interval}\n" + \
             f"\n" + \
             f"cluster:\n" + \
             f"  node: 10\n" + \
             f"\n" + \
             f"load-balancer:\n" + \
             f"  algo: {algo}\n" + \
             f"  local-port: 8000\n" + \
             f"\n" + \
             f"experiment:\n" + \
             f"  name: {name}\n" + \
             f"  latency: {latency}\n"+ \
             f"  homogeneous: {str(homog).lower()}"
    return config

def get_requests(rate, interval, sz):
    amount = rate * interval
    return [amount * int(sz == "s"), amount * int(sz == "m"), amount * int(sz == "l")]

def generate_configs():
    for algo in ALG_OPTIONS:
        for latency in LATENCY_OPTIONS:
            for homog in HOMOG_OPTIONS:
                for f_sz in FILE_SZ:
                    for rate in RATES: # requests/sec
                        if ((latency == 0 and homog == True and f_sz == "m") or
                              latency == 100 and homog == False and f_sz == "m"):
                            name = f"exp-{algo[0]}-lat-{latency}-homog-{str(homog).lower()}-int-{INTERVAL}-fsz-{f_sz}-rate-{rate}"
                            config = gen_config(name, algo[1], homog, latency, INTERVAL, get_requests(rate, INTERVAL, f_sz))
                            fd = open(f"./config/{name}.yml", "w")
                            fd.write(config)
                            fd.close()
    return

def get_client_exp( filename ):
    pattern = re.compile(rf"{USER_DIR}/client-exp-(rr|lc|lrt)-lat-(\d+)-homog-(true|false)-int-(\d+)-fsz-(s|m|l)-rate-(\d+)\.csv")
    match = pattern.match(filename)
    if not match:
        print(f"Error: unable to match file {filename}")
        return ClientExp("ERROR")

    alg, net_delay, homog, interval, fsz, rate = match.groups()
    sz = 0

    try:
        fd = open(filename, "r")
    except:
        print(f"Error attempting to open file {filename}")
        sys.exit(1)

    # Read the header
    fd.readline()

    total = 0
    length = 0

    # Save data in dict
    for line in fd:
        dur, sz = line.split(",")
        dur = float(dur)
        sz = int(sz)
        if dur == 0 and sz == 0:
            continue # skip failed requests
        total += float(dur) / 1000
        length += 1
    
    fd.close()

    if (length == 0):
        print(f'File {filename} had all zero entries')
        return ClientExp("ERROR")
    
    homog = True if homog == "true" else False

    return ClientExp(alg, int(net_delay), homog, int(interval), fsz, int(sz), int(rate), total/length)

def generate_client_avg_time_vs_size(data, title, figure_name) -> None:
    # Save the plot in the plots dir
    save_dir = "plots"
    os.makedirs(save_dir, exist_ok=True)

    # Sort bins by request rate
    {data[bin].sort(key=lambda c: c.rate) for bin in data}

    # Plot
    plt.figure(figsize=(8, 5))
    for bin in data:
        series = data[bin]
        avg_times = [c.avg_serv_time for c in series]
        plt.plot(RATES, avg_times, marker='o', linestyle='-', label=bin.upper())

    plt.xscale("log", base=10)
    plt.xlabel('Request Rate (requests/sec)')
    plt.ylabel('Average Service Time (sec)')
    plt.title(title)
    plt.legend()
    plt.savefig(os.path.join(save_dir, figure_name))

    return

def generate_client_experiment_records(files):
    records = []
    for file in files:
        records.append(get_client_exp(file))
    
    return records

def filter_client_records(records, filter):
    return [r for r in records if filter(r)]

def get_filenames_from_dir(directory):
    dir_files = []
    try:
        dir_files = os.listdir(directory)
    except FileNotFoundError:
        print(f"Directory {directory} not found")
        sys.exit(1)

    files = []
    for filename in dir_files:
        files.append(os.path.join(directory, filename))
    return files

def get_homog_string(homog):
    return "Homogeneous" if homog else "Heterogeneous"

def get_file_size_string(sz):
    if sz == "s":
        return "Small"
    elif sz == "m":
        return "Medium"
    else:
        return "Large"
    
def get_alg_string(alg):
    if alg == "rr":
        return "Round Robin"
    elif alg == "lc":
        return "Least Connections"
    else:
        return "Least Service Time"

def get_latency_string(latency):
    if latency == 0:
        return "Zero Network Delay"
    else:
        return f"{latency}ms Network Delay"

def bin_data_by_alg(data):
    bins = {"rr": [], "lc": [], "lrt": []}
    for r in data:
        bins[r.alg].append(r)
    return bins

def bin_data_by_fsz(data):
    bins = {"s": [], "m": [], "l": []}
    for r in data:
        bins[r.fsz].append(r)
    return bins

def generate_user_plots():
    files = get_filenames_from_dir(USER_DIR)
    exp_records = generate_client_experiment_records(files)

    # Charts with Avg time on Y, Request Rate on X, Same file Size, Same Lat, Same Homog, All Algs
    for homog in HOMOG_OPTIONS:
        for sz in FILE_SZ:
            for net_delay in LATENCY_OPTIONS:
                filter = lambda r: r.homog == homog and r.fsz == sz and r.net_delay == net_delay
                title = f"Average Client Request Service Time for {get_file_size_string(sz)} Files with Various"
                title += f"\nRequest Rates with {get_homog_string(homog)} Nodes and {get_latency_string(net_delay)}"
                figure_name = f"algs-compare-homog-{homog}-fsz-{sz}-delay-{net_delay}"
                filtered_records = filter_client_records(exp_records, filter)
                binned_data = bin_data_by_alg(filtered_records)
                if all(len(binned_data[b]) == 3 for b in binned_data):
                    generate_client_avg_time_vs_size(binned_data, title, figure_name)
    
    # Charts with Avg time on Y, Request Rate on X, Different File Sizes, Same Lat, Same Homog, Same Alg
    for homog in HOMOG_OPTIONS:
        for alg, _ in ALG_OPTIONS:
            for net_delay in LATENCY_OPTIONS:
                filter = lambda r: r.homog == homog and r.alg == alg and r.net_delay == net_delay
                title = f"{get_alg_string(alg)}: Service Time vs. Request Size"
                figure_name = f"vary-fsz-alg-{alg}-homog-{homog}-delay-{net_delay}"
                filtered_records = filter_client_records(exp_records, filter)
                binned_data = bin_data_by_fsz(filtered_records)
                if all(len(binned_data[b]) == 3 for b in binned_data):
                    generate_client_avg_time_vs_size(binned_data, title, figure_name)


def main():
    if (len(sys.argv) != 2):
        print("Error: invalid number of command arguments")
        print("Usage: `python3 gen_files.py <opt>` where <opt> can be one of `user` or `configs`")
        sys.exit(1)
    
    if sys.argv[1] == "user":
        generate_user_plots()
    elif sys.argv[1] == "configs":
        generate_configs()
    else:
        print("Error: invalid option given")
        print("Usage: `python3 gen_files.py <opt>` where opt can be one of `user` or `configs`")
        sys.exit(1)


if __name__ == "__main__":
    main()
