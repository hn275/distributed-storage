# distributed-storage

A reseach project, focusing on perfomance analysis of various load balancing
techniques in a P2P system.

For details of our analysis, see [report.pdf](./report.pdf).

# Simulation

## File Data

If you're running the simulation in a Docker environment, this step can be
safely skipped. Proceed directly to the [Docker Container](#docker-container)
section.

For bare-metal execution, the simulation data must first be generated using
the `data-gen` executable:

```sh
go run ./cmd/data-gen
```

This will create a `./tmp/data` directory and populate it with six files of
varying sizes used in the simulation. The entire `./tmp` directory is included
in `.gitignore`, as the generated data can total up to 1 GB.

## Bare Metal

Bare-metal execution **is only supported on Unix environment**, and is not
recommended due to the additional tooling required. At minimum, compatible Go
compiler and Python interpreter must be installed and properly configured. See
[Environment](#environment) for details.

### Go Module

The required binaries must be compiled and placed in `./tmp/bin/`. This can be
done using the provided `make` rule:

```sh
make compile
```

As noted in the section [File Data](#file-data), ensure the mock data is
generated before running the simulation.

### Python Module

To generate graphs and perform the necessary analysis, install the Python
dependencies:

```sh
pip install -r requirements.txt
```

### Execution

With the binaries compiled and Python modules installed, the simulation can be
executed using:

```sh
./runsim.sh
```

## Docker Container

While the simulation can be run manually using `./runsim.sh`, we recommend
using Docker and Docker Compose to automate code compilation, dependency
management, simulation execution, and network topology initialization.

An optional environment variable `LATENCY` can be passed in to emulate network
latency, the value is a integer in the unit of milliseconds. For example-to
run the simulation with 25ms of network latency:

```sh
LATENCY=25 docker compose up --build
```

> **Note:** Use the `--build` flag only after making any code changes to
> trigger recompilation.

Simulation output is mounted to two directories: logs are stored in `./log/`,
and telemetry data is saved in `./output/`. Depending on your system’s Docker
group permissions, you may need to change ownership of these directories to
access the output:

```sh
sudo chown <owner>:<group> -R ./output/ ./log/
```

# Development

## Environment

This project was developed and tested using the following versions:

- Go 1.24.2
- Python 3.13.3
- GNU Make 4.4.1

## Project Directory Structure

| Type             | Path                         |
| :--------------- | :--------------------------- |
| Executables      | `./cmd/<binary>/main.go`     |
| Private packages | `./internal/<binary-name>/`  |
| Shared packages  | `./internal/<package-name>/` |

## CI/CD

Two pipelines are configured to ensure dependency validity and consistent code
formatting. Both must pass before any pull request can be merged into the
`main` branch.
