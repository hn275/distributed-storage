#!/usr/bin/python3

ALG_OPTIONS = (("rr", "simple-round-robin"), ("lc", "least-connections"), ("lrt", "least-response-time"))
HOMOG_OPTIONS = (True, False)
INTERVAL = 10
FILE_SZ = ("s", "m", "l")
RATES = (10, 100, 1000)
LATENCY_OPTIONS = (0, 100)

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
                        name = f"exp-{algo[0]}-lat-{latency}-homog-{str(homog).lower()}-int-{INTERVAL}-fsz-{f_sz}-rate-{rate}"
                        config = gen_config(name, algo[1], homog, latency, INTERVAL, get_requests(rate, INTERVAL, f_sz))
                        fd = open(f"./config/{name}.yml", "w")
                        fd.write(config)
                        fd.close()
    return
    

     



def main():
    generate_configs()

if __name__ == "__main__":
    main()
