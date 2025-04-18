#!/usr/bin/python3

import re
import sys
import matplotlib.pyplot as plt
import os

LB_EVENT_IDX = 0
LB_TIME_IDX = 3
LB_QUEUE_IDX = 6
LB_NODE_IDX = 2
CLUSTER_SIZE = 20

CL_EVENT_IDX = 2
CL_NODE_IDX = 0
CL_NUM_CONN = 8
CL_TIME_IDX = 5

save_dir = "tmp/output/investigation"


def open_file(filename):
    try:
        fd = open(filename, "r")
    except:
        print(f"Error: unable to open file {filename}")
        sys.exit(1)

    return fd


def generate_cl_num_conn_distribution(filename, output_file):
    fd = open_file(filename)
    time_label = "times"
    conn_label = "conn"

    node_info = {}
    for idx in range(CLUSTER_SIZE):
        node_info[idx] = {time_label: [], conn_label: []}

    fd.readline()  # read the header

    for line in fd:
        tokens = line.split(",")
        if tokens[CL_EVENT_IDX] == "healthcheck":
            time = int(tokens[CL_TIME_IDX])
            node_id = int(tokens[CL_NODE_IDX])
            num_conn = int(tokens[CL_NUM_CONN])
            node_info[node_id][time_label].append(time)
            node_info[node_id][conn_label].append(num_conn)

    fd.close()  # Done reading so close file

    # Plot number of connections over time
    plt.figure(figsize=(8, 5))
    colors = plt.get_cmap("tab20").colors
    for idx, node_id in enumerate(node_info):
        plt.plot(
            node_info[node_id][time_label],
            node_info[node_id][conn_label],
            marker="o",
            linestyle="-",
            label=str(node_id),
            color=colors[idx],
        )

    plt.xlabel("Time (ns)")
    plt.ylabel("Number of Connections")
    plt.title("Cluster: Number of Connections over Time")
    plt.legend()
    plt.savefig(os.path.join(save_dir, "cl-queue-" + output_file), dpi=300)
    plt.close()


def generate_lb_num_conn_distribution(filename, output_file):
    fd = open_file(filename)

    fd.readline()  # read the header

    node_to_conn = {}
    times = []
    for idx in range(CLUSTER_SIZE):
        node_to_conn[idx] = []

    for line in fd:
        tokens = line.split(
            ",", maxsplit=6
        )  # TODO: may need to change this val depending on csv format
        event = tokens[LB_EVENT_IDX]

        if event == "user-joined":
            time = int(tokens[LB_TIME_IDX])
            queue = tokens[LB_QUEUE_IDX]

            queue_tuples = re.findall(r"\((\d+,\d+)\)", queue)

            assert len(queue_tuples) == CLUSTER_SIZE

            # Extract number of connections for each node in queue and store in dict
            for e in queue_tuples:
                node_id, num_conn = e.split(",")
                node_id = int(node_id)
                num_conn = int(num_conn)
                node_to_conn[node_id].append(num_conn)

            times.append(time)

    fd.close()  # Done reading so close file

    # Plot number of connections over time
    plt.figure(figsize=(8, 5))
    colors = plt.get_cmap("tab20").colors
    for idx, node_id in enumerate(node_to_conn):
        plt.plot(
            times,
            node_to_conn[node_id],
            marker="o",
            linestyle="-",
            label=str(node_id),
            color=colors[idx],
        )

    plt.xlabel("Time (ns)")
    plt.ylabel("Number of Connections")
    plt.title("LB: Number of Connections over Time")
    plt.legend()
    plt.savefig(os.path.join(save_dir, "lb-queue-" + output_file), dpi=300)
    plt.close()


def filter_lb_logs(filename, output_file, node_filter=None, time_filter=None):
    fd_r = open_file(filename)
    fd_w = open(os.path.join(save_dir, "filtered-" + output_file), "w")
    # write the header to the file
    fd_w.write(fd_r.readline())

    for line in fd_r:
        tokens = line.split(",")
        event = tokens[LB_EVENT_IDX]
        node_id = int(tokens[LB_NODE_IDX])
        timestamp = int(tokens[LB_TIME_IDX])
        if (
            event == "user-joined"
            and (node_filter == None or node_filter == node_id)
            and (time_filter == None or time_filter <= timestamp)
        ):
            fd_w.write(line)

    fd_r.close()
    fd_w.close()


def filter_cluster_logs(filename, output_file, node_filter=None, time_filter=None):
    fd_r = open_file(filename)
    fd_w = open(os.path.join(save_dir, "filtered-" + output_file), "w")
    # write the header to the file
    fd_w.write(fd_r.readline())

    for line in fd_r:
        tokens = line.split(",")
        event = tokens[CL_EVENT_IDX]
        node_id = int(tokens[CL_NODE_IDX])
        timestamp = int(tokens[CL_TIME_IDX])
        if (
            event == "healthcheck"
            and (node_filter == None or node_filter == node_id)
            and (time_filter == None or time_filter <= timestamp)
        ):
            fd_w.write(line)

    fd_r.close()
    fd_w.close()


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Error: need at least an option and a filename")
        print(
            """
Usage Variants:
- python3 inv-script.py <filter-lb|fileter-cl> <input file>
- python3 inv-script.py <filter-lb|fileter-cl> <input file> n=<node-id> t=<time in ns>
- python3 inv-script.py <plot-cl|plot-lb> <input file>
"""
        )
        sys.exit(1)

    option = sys.argv[1]

    os.makedirs(save_dir, exist_ok=True)

    input_file = sys.argv[2]
    output_file = input_file.split("/")[3]

    if option == "filter-cl" or option == "filter-lb":
        node = None
        time = None
        if len(sys.argv) == 5:
            node_arg = sys.argv[3]
            _, id = node_arg.split("=")
            node = int(id)

            time_arg = sys.argv[4]
            _, t = time_arg.split("=")
            time = int(t)

        if option == "filter-cl":
            filter_cluster_logs(
                input_file, output_file, node_filter=node, time_filter=time
            )
        if option == "filter-lb":
            filter_lb_logs(input_file, output_file, node_filter=node, time_filter=time)

    elif option == "plot-cl":

        output_file = output_file.split(".")[0]
        generate_cl_num_conn_distribution(input_file, output_file)

    elif option == "plot-lb":
        output_file = output_file.split(".")[0]
        generate_lb_num_conn_distribution(input_file, output_file)

    else:
        print(
            "Invalid option given. Must be one of filter-cl, filter-lb, plot-cl, plot-lb"
        )
