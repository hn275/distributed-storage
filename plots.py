import sys
import matplotlib.pyplot as plt
import numpy as np
import os


def main():
    if len(sys.argv) < 3:
        print("Invalid number of arguments")
        sys.exit(1)

    _, exp_type, filename = sys.argv

    if exp_type == "CLIENT-DATA":
        generate_client_avg_time_vs_size(filename)

    return


def generate_client_avg_time_vs_size(filename: str) -> None:
    try:
        fd = open(filename, "r")
    except:
        print(f"Error attempting to open file {filename}")
        sys.exit(1)

    size_to_time = {}

    # Read the header
    fd.readline()

    # Save data in dict
    for line in fd:
        dur, size = line.split(",")
        dur = float(dur) / 1000
        size = int(size)

        if size not in size_to_time:
            size_to_time[size] = [dur]
        else:
            size_to_time[size].append(dur)

    save_dir = "plots"
    os.makedirs(save_dir, exist_ok=True)

    # Compute average times for each size
    sizes_sorted = sorted(list(size_to_time.keys()))
    average_times = [np.mean(size_to_time[size]) for size in sizes_sorted]

    # Plot
    plt.figure(figsize=(8, 5))
    plt.plot(sizes_sorted, average_times, marker="o", linestyle="-")
    plt.xscale("log", base=2)
    plt.xlabel("File Size (Bytes)")
    plt.ylabel("Average Service Time (s)")
    plt.title("Average Client Request Service Type for Files of Various Sizes")
    plt.savefig(os.path.join(save_dir, "experiment-results"))

    return


if __name__ == "__main__":
    main()
