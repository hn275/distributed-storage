#!/usr/bin/python3

import re
import sys
import matplotlib.pyplot as plt
import os
import random
import matplotlib as mpl

mpl.rcParams.update({"font.size": 12})  # Set global font size

ALG_OPTIONS = (
    ("rr", "simple-round-robin"),
    ("lc", "least-connections"),
    ("lrt", "least-response-time"),
)
HOMOG_OPTIONS = (True, False)
INTERVAL = 20
FILE_SZ = ("s", "m", "l", "v")  # be sure that v is always the final element
RATES = (10, 50, 100, 150, 200)
LATENCY_OPTIONS = (0, 25)
USER_DIR = "tmp/output/user"
LB_DIR = "tmp/output/lb"
CLUSTER_DIR = "tmp/output/cluster"
DATANODE_COUNT = 10
CLUSTER_CAPACITY = 5


class ClientExp:
    def __init__(
        self,
        alg,
        fname=None,
        net_delay=None,
        homog=None,
        interval=None,
        fsz=None,
        fsz_bytes=None,
        rate=None,
        avg_serv_time=None,
        errors=None,
        values=None,
    ):
        self.alg = alg
        self.fname = fname
        self.net_delay = net_delay
        self.homog = homog
        self.interval = interval
        self.fsz = fsz
        self.fsz_bytes = fsz_bytes
        self.rate = rate
        self.avg_serv_time = avg_serv_time
        self.errors = errors  # errors as a percentages
        self.values = values

    def __repr__(self):
        return f"ClientExp(alg={self.alg}, net_delay={self.net_delay}, homog={self.homog}, interval={self.interval}, fsz={self.fsz}, rate={self.rate}, avg_serv_time={self.avg_serv_time}, errors={self.errors}, std_dev={self.std_dev}"


class LBExp:
    def __init__(
        self,
        alg,
        net_delay=None,
        homog=None,
        interval=None,
        fsz=None,
        rate=None,
        req_per_sec=None,
    ):
        self.alg = alg
        self.net_delay = net_delay
        self.homog = homog
        self.interval = interval
        self.fsz = fsz
        self.rate = rate
        self.req_per_sec = req_per_sec

    def __repr__(self):
        return f"LBExp(alg={self.alg}, net_delay={self.net_delay}, homog={self.homog}, interval={self.interval}, fsz={self.fsz}, rate={self.rate}, avg_serv_time={self.req_per_sec}"


def gen_config(name, algo, homog, latency, interval, files):
    config = (
        f"user:\n"
        + f"  small: {files[0]}\n"
        + f"  medium: {files[1]}\n"
        + f"  large: {files[2]}\n"
        + f"  x-large: 0\n"
        + f"  xx-large: 0\n"
        + f"  interval: {interval}\n"
        + f"\n"
        + f"cluster:\n"
        + f"  node: {DATANODE_COUNT}\n"
        + f"  capacity: {CLUSTER_CAPACITY}\n"
        + f"\n"
        + f"load-balancer:\n"
        + f"  algo: {algo}\n"
        + f"  local-port: 8000\n"
        + f"\n"
        + f"experiment:\n"
        + f"  name: {name}\n"
        + f"  latency: {latency}\n"
        + f"  homogeneous: {str(homog).lower()}\n"
        + f"  overhead-param: 1000"
    )  # change this, it's in nanoseconds
    return config


def get_requests(rate, interval, sz):
    amount = rate * interval
    return [amount * int(sz == "s"), amount * int(sz == "m"), amount * int(sz == "l")]


def generate_varying_amounts_per_rate():
    amounts = {}
    for rate in RATES:
        amount = rate * INTERVAL
        # make a bucket for each non "v" file size
        amounts[rate] = [0 for _ in range(len(FILE_SZ) - 1)]
        random.seed(42)  # use seed to ensure the same numbers are always generated
        for _ in range(amount):
            index = random.randint(
                0, len(FILE_SZ) - 2
            )  # generate a random index (that is not for "v")
            amounts[rate][index] += 1

    return amounts


def generate_configs(opt):
    varying_fsz_with_fixed_amount = generate_varying_amounts_per_rate()

    for algo in ALG_OPTIONS:
        for latency in LATENCY_OPTIONS:
            for homog in HOMOG_OPTIONS:
                for f_sz in FILE_SZ:
                    for rate in RATES:  # requests/sec
                        if (
                            latency == LATENCY_OPTIONS[0]
                            and homog == False
                            and f_sz == "v"
                            and opt == "configs1"
                        ) or (
                            latency == LATENCY_OPTIONS[1]
                            and f_sz == "v"
                            and opt == "configs2"
                        ):
                            name = f"exp-{algo[0]}-lat-{latency}-homog-{str(homog).lower()}-int-{INTERVAL}-fsz-{f_sz}-rate-{rate}"
                            requests = (
                                varying_fsz_with_fixed_amount[rate]
                                if f_sz == "v"
                                else get_requests(rate, INTERVAL, f_sz)
                            )
                            config = gen_config(
                                name, algo[1], homog, latency, INTERVAL, requests
                            )
                            config_path = f"./config/{name}.yml"
                            fd = open(config_path, "w")
                            fd.write(config)
                            fd.close()
                            print(f"Config generated: {config_path}")
    return


def get_client_exp(filename):
    pattern = re.compile(
        rf"{USER_DIR}/client-(exp-(rr|lc|lrt)-lat-(\d+)-homog-(true|false)-int-(\d+)-fsz-(s|m|l|v)-rate-(\d+))\.csv"
    )
    match = pattern.match(filename)
    if not match:
        print(f"Error: unable to match file {filename}")
        return ClientExp("ERROR")

    fname, alg, net_delay, homog, interval, fsz, rate = match.groups()
    sz = 0
    values = []

    try:
        fd = open(filename, "r")
    except:
        print(f"Error attempting to open file {filename}")
        sys.exit(1)

    # Read the header
    fd.readline()

    total = 0
    length_no_errors = 0
    errors = 0
    full_length = 0

    # Save data in dict
    for line in fd:
        items = line.split(",")
        dur = items[0]
        sz = items[1]

        dur = float(dur)
        sz = int(sz)
        if dur == 0 and sz == 0:
            errors += 1

        else:
            dur = dur / 1000
            total += dur
            length_no_errors += 1
            values.append(dur)

        full_length += 1

    fd.close()

    if length_no_errors == 0:
        print(f"File {filename} had all zero entries")
        return ClientExp("ERROR")

    homog = True if homog == "true" else False

    return ClientExp(
        alg,
        fname,
        int(net_delay),
        homog,
        int(interval),
        fsz,
        int(sz),
        int(rate),
        total / length_no_errors,
        errors / full_length,
        values,
    )


def generate_client_serv_time_distribution(client_exp, figure_name):
    save_dir = "tmp/output/client-distributions"
    os.makedirs(save_dir, exist_ok=True)

    plt.hist(client_exp.values, 50, color="skyblue")
    plt.title(f"{get_alg_string(client_exp.alg)}: Service Time Distribution", pad=20)
    plt.ylabel("Frequency")
    plt.xlabel("Service Time (sec)")
    plt.savefig(os.path.join(save_dir, figure_name), dpi=300, bbox_inches="tight")
    plt.close()


def get_lb_exp(filename):
    pattern = re.compile(
        rf"{LB_DIR}/lb-exp-(rr|lc|lrt)-lat-(\d+)-homog-(true|false)-int-(\d+)-fsz-(s|m|l|v)-rate-(\d+)\.csv"
    )
    match = pattern.match(filename)
    if not match:
        print(f"Error: unable to match file {filename}")
        return ClientExp("ERROR")

    alg, net_delay, homog, interval, fsz, rate = match.groups()

    try:
        fd = open(filename, "r")
    except:
        print(f"Error attempting to open file {filename}")
        sys.exit(1)

    # Read the header
    fd.readline()

    total_req = 0
    start_first_request = None
    end_last_request = 0

    # Save data in dict
    for line in fd:
        event, _, _, start_time, dur = line.split(",")
        if event == "user-joined":
            start_time = float(start_time)
            end_time = start_time + float(dur)
            if start_first_request == None:
                start_first_request = start_time

            if end_last_request < end_time:
                end_last_request = end_time

            total_req += 1

    fd.close()

    if total_req == 0:
        print(f"File {filename} had no entries with event-type user-joined")

    homog = True if homog == "true" else False

    req_per_sec = 0
    if total_req > 0:
        total_time = (
            end_last_request - start_first_request
        ) / 1e9  # convert total time to seconds
        req_per_sec = total_req / total_time

    return LBExp(alg, int(net_delay), homog, int(interval), fsz, int(rate), req_per_sec)


def generate_client_avg_time_vs_size(data, title, figure_name) -> None:
    # Save the plot in the plots dir
    save_dir = "tmp/output/plots"
    os.makedirs(save_dir, exist_ok=True)

    # Sort bins by request rate
    {data[bin].sort(key=lambda c: c.rate) for bin in data}
    END = len(RATES)
    # Plot
    plt.figure(figsize=(8, 5))
    colors = plt.get_cmap("tab10").colors
    for idx, bin in enumerate(data):
        series = data[bin]
        avg_times = [c.avg_serv_time for c in series]
        plt.plot(
            RATES[:END],
            avg_times[:END],
            marker="o",
            linestyle="-",
            label=bin.upper(),
            color=colors[idx],
        )

    plt.xlabel("Request Rate (requests/sec)")
    plt.ylabel("Average Service Time (sec)")
    plt.title(title)
    plt.legend()
    plt.savefig(os.path.join(save_dir, figure_name), dpi=300)

    return


def generate_lb_table_results(binned_data, output_file, bin_label):
    save_dir = "lb_tables"
    os.makedirs(save_dir, exist_ok=True)

    try:
        fd = open(os.path.join(save_dir, output_file), "w")
    except:
        print(f"Error attempting to open file {dir}{output_file} for writing")
        sys.exit(1)

    header = f"{bin_label},"
    for i in range(len(RATES)):
        header += str(RATES[i]) + "(req/sec)"
        if i < len(RATES) - 1:
            header += ","
    fd.write(f"{header}\n")

    # Sort bins by request rate
    {binned_data[b].sort(key=lambda lb: lb.rate) for b in binned_data}

    # Fill in rows of the csv
    for bucket in binned_data:
        row = f"{bucket},"
        for i in range(len(binned_data[bucket])):
            row += str(binned_data[bucket][i].req_per_sec)
            if i < len(binned_data[bucket]) - 1:
                row += ","
        fd.write(f"{row}\n")

    fd.close()


def generate_experiment_records(files, item_generator):
    records = []
    for file in files:
        records.append(item_generator(file))

    return records


def filter_records(records, filter):
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
    sz_to_string = {"s": "Small", "m": "Medium", "l": "Large", "v": "Varied-sized"}
    return sz_to_string[sz]


def get_alg_string(alg):
    if alg == "rr":
        return "Round Robin"
    elif alg == "lc":
        return "Least Connections"
    else:
        return "Least Response Time"


def get_latency_string(latency):
    if latency == 0:
        return "Zero Network Delay"
    else:
        return f"25 ms Network Delay"


def bin_data_by_alg(data):
    bins = {"rr": [], "lc": [], "lrt": []}
    for r in data:
        bins[r.alg].append(r)
    return bins


def bin_data_by_fsz(data):
    bins = {"s": [], "m": [], "l": []}
    for r in data:
        if r.fsz != "v":
            bins[r.fsz].append(r)
    return bins


def generate_client_errors(binned_data, output_file, bin_label):
    save_dir = "tmp/output/client-errors"
    os.makedirs(save_dir, exist_ok=True)

    try:
        fd = open(os.path.join(save_dir, output_file + ".csv"), "w")
    except:
        print(f"Error attempting to open file {dir}{output_file} for writing")
        sys.exit(1)

    header = f"{bin_label},"
    for i in range(len(RATES)):
        header += str(RATES[i]) + "(req/sec)"
        if i < len(RATES) - 1:
            header += ","
    fd.write(f"{header}\n")

    # Sort bins by request rate
    {binned_data[b].sort(key=lambda c: c.rate) for b in binned_data}

    # Fill in rows of the csv
    for bucket in binned_data:
        row = f"{bucket},"
        for i in range(len(binned_data[bucket])):
            row += str(binned_data[bucket][i].errors)
            if i < len(binned_data[bucket]) - 1:
                row += ","
        fd.write(f"{row}\n")

    fd.close()


def generate_user_plots():
    files = get_filenames_from_dir(USER_DIR)
    exp_records = generate_experiment_records(files, get_client_exp)

    # Generate user service time distributions
    for c in exp_records:
        if c.alg != "ERROR":
            figure_name = "hist-" + c.fname
            generate_client_serv_time_distribution(c, figure_name)
            print(f"Histogram generated: {figure_name}")

    # Charts with Avg time on Y, Request Rate on X, Same file Size, Same Lat, Same Homog, All Algs
    for homog in HOMOG_OPTIONS:
        for sz in FILE_SZ:
            for net_delay in LATENCY_OPTIONS:
                filter = (
                    lambda r: r.homog == homog
                    and r.fsz == sz
                    and r.net_delay == net_delay
                )
                title = f"Average Client Request Service Time vs Request Rate"
                figure_name = f"algs-compare-homog-{homog}-fsz-{sz}-delay-{net_delay}"
                filtered_records = filter_records(exp_records, filter)
                binned_data = bin_data_by_alg(filtered_records)
                # Each bin should have an entry for each rate in order for this to be a valid configuration
                if all(len(binned_data[b]) == len(RATES) for b in binned_data):
                    generate_client_avg_time_vs_size(binned_data, title, figure_name)
                    generate_client_errors(binned_data, figure_name, "Algorithm")
                    print(f"Plot generated: {figure_name}")

    # Charts with Avg time on Y, Request Rate on X, Different File Sizes, Same Lat, Same Homog, Same Alg
    for homog in HOMOG_OPTIONS:
        for alg, _ in ALG_OPTIONS:
            for net_delay in LATENCY_OPTIONS:
                filter = (
                    lambda r: r.homog == homog
                    and r.alg == alg
                    and r.net_delay == net_delay
                )
                title = f"{get_alg_string(alg)}: Service Time vs. Request Size"
                figure_name = f"vary-fsz-alg-{alg}-homog-{homog}-delay-{net_delay}"
                filtered_records = filter_records(exp_records, filter)
                binned_data = bin_data_by_fsz(filtered_records)
                # Each bin should have an entry for each rate in order for this to be a valid configuration
                if all(len(binned_data[b]) == len(RATES) for b in binned_data):
                    generate_client_avg_time_vs_size(binned_data, title, figure_name)
                    generate_client_errors(binned_data, figure_name, "File Size")
                    print(f"Plot generated: {figure_name}")


def generate_lb_data():
    files = get_filenames_from_dir(LB_DIR)
    exp_records = generate_experiment_records(files, get_lb_exp)

    # Tables with cols for alg, 10, 32, 100, 320, 1000 rate
    for homog in HOMOG_OPTIONS:
        for sz in FILE_SZ:
            for net_delay in LATENCY_OPTIONS:
                filter = (
                    lambda r: r.homog == homog
                    and r.fsz == sz
                    and r.net_delay == net_delay
                )
                output_file = (
                    f"algs-compare-homog-{homog}-fsz-{sz}-delay-{net_delay}.csv"
                )
                filtered_records = filter_records(exp_records, filter)
                binned_data = bin_data_by_alg(filtered_records)
                # Each bin should have an entry for each rate in order for this to be a valid configuration
                if all(len(binned_data[b]) == len(RATES) for b in binned_data):
                    generate_lb_table_results(binned_data, output_file, "Algorithm")

    # Tables with cols for file size, 10, 32, 100, 320, 1000 rate
    for homog in HOMOG_OPTIONS:
        for alg, _ in ALG_OPTIONS:
            for net_delay in LATENCY_OPTIONS:
                filter = (
                    lambda r: r.homog == homog
                    and r.alg == alg
                    and r.net_delay == net_delay
                )
                output_file = f"vary-fsz-alg-{alg}-homog-{homog}-delay-{net_delay}.csv"
                filtered_records = filter_records(exp_records, filter)
                binned_data = bin_data_by_fsz(filtered_records)
                # Each bin should have an entry for each rate in order for this to be a valid configuration
                if all(len(binned_data[b]) == len(RATES) for b in binned_data):
                    generate_lb_table_results(binned_data, output_file, "File Size")


def generate_requests_per_node():
    files = get_filenames_from_dir(CLUSTER_DIR)

    save_dir = "tmp/output/node-request-counts"
    os.makedirs(save_dir, exist_ok=True)

    for file in files:
        pattern = re.compile(
            rf"{CLUSTER_DIR}/cluster-exp-(((rr|lc|lrt)-lat-(?:\d+)-homog-(?:true|false)-int-(?:\d+)-fsz-(?:s|m|l|v)-rate-(?:\d+))\.csv)"
        )
        match = pattern.match(file)

        if match == None:
            continue

        output_file = match.groups()[0]
        node_req_count = {}

        try:
            fd = open(file, "r")
        except:
            print(f"Error attempting to open file {CLUSTER_DIR}/{file} for reading")
            sys.exit(1)

        # node-id,performance-overhead(ns),event-type,peer,timestamp,duration(ns),bytes-transferred
        fd.readline()  # header

        overhead_map = {}
        # read all lines in file and gather node request counts
        for line in fd:
            id, overhead, event, _ = line.split(",", maxsplit=3)
            # overhead = int(overhead)
            id = int(id)
            if event == "file-transfer":
                if id not in node_req_count:
                    node_req_count[id] = 1
                else:
                    node_req_count[id] += 1
                overhead_map[id] = overhead

        fd.close()

        # write the results to output file
        try:
            fd = open(os.path.join(save_dir, "node-req-count-" + output_file), "w")
        except:
            print(f"Error attempting to open file {dir}/{output_file} for writing")
            sys.exit(1)

        fd.write("node-id,count,overhead\n")

        sorted_ids = sorted(list(node_req_count.keys()))
        values = []
        for id in sorted_ids:
            values.append(node_req_count[id])

        sorted_overhead = sorted(list(overhead_map.items()), key=lambda tup: tup[1])
        for id, _ in sorted_overhead:
            fd.write(f"{id},{node_req_count[id]},{overhead_map[id]}\n")

        fd.close()

        # bar chart
        figure_name = os.path.join(save_dir, "node-req-count-" + match.groups()[1])
        plt.figure()
        plt.bar(sorted_ids, values, color="skyblue")
        plt.xlabel("Node ID")
        plt.ylabel("Number of Requests")
        plt.xticks(sorted_ids)
        plt.title(f"{get_alg_string(match.groups()[2])}: Requests per Node")
        plt.savefig(figure_name, dpi=300, bbox_inches="tight")
        plt.close()
        print(f"Output generated: {figure_name}")
    return


def main():
    if len(sys.argv) != 2:
        print("Error: invalid number of command arguments")
        print(
            "Usage: `python3 gen_files.py <opt>` where <opt> can be one of `user`, `cluster`, `configs1`, or `configs2`"
        )
        print(
            "Note: To generate configs used in project experiments, use `configs2` option"
        )
        sys.exit(1)

    if sys.argv[1] == "user":
        generate_user_plots()
    elif sys.argv[1] == "configs1" or sys.argv[1] == "configs2":
        generate_configs(sys.argv[1])
    elif sys.argv[1] == "cluster":
        generate_requests_per_node()
    else:
        print("Error: invalid option given")
        print(
            "Usage: `python3 gen_files.py <opt>` where opt can be one of `user`, `cluster`, `configs1`, or `configs2`"
        )
        print(
            "Note: To generate configs used in project experiments, use `configs2` option"
        )
        sys.exit(1)


if __name__ == "__main__":
    main()
