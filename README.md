# distributed-storage

An ongoing reseach project, focusing on perfomance analysis of various load
balancing techniques in a P2P system.

# Simulation

For the simulation, we will use Docker and docker-compose to automate the code
compilation and initialization of the required network topology processes.

TODO: redo docker compose and update the docs

# Development

## File Data

Before running the simulation, the file data can be generated with the
`data-gen` executable:

```sh
go run ./cmd/data-gen
```

This script will create a dir entry `tmp/data`, and 6 files of different size
for the simulation. The entire `tmp` directory is added to `.gitignore`, since
the generated files can be as large as 1Gb.

## Project Directory Structure

| Type             | Path                         |
| :--------------- | :--------------------------- |
| Executables      | `./cmd/<binary>/main.go`     |
| Private packages | `./internal/<binary-name>/`  |
| Shared packages  | `./internal/<package-name>/` |

## CI/CI: Code Validation

A [CI/Code Validation](./.github/workflows/ci.yml) pipeline is set up for code
validation when a PR is opened. The action is required to complete without
errors before the PR can be pulled into `main`.

### Code Formatting

In the case the the action failed because your code isn't formatted, use `gofmt`

```sh
# to format all files
gofmt -w .

# or to format a specific file
gofmt -w ./path/to/file
```

For more usage, see the [docs](https://pkg.go.dev/cmd/gofmt) for `gofmt`.

# Requirements

- User Interface
  - simple, hard code binary to simulate sending an arbituary number requests.
- Load Balancing
  - supported multiple algorithms, but start with one
  - have an interface so different algorithms can just be plugged and play
    - can be stateful, different algos can have different struct type, etc etc
    - internal state processing
- Data Node
  - no communication needed between data nodes.
  - all nodes have the same data.
  - directly connect to the users for the data transfer.
  - encryption added, but set with a flag.
  - updating states to LB.
  - can simulate heterogeneous servers with `sleep`.
