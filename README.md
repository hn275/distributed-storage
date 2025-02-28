# distributed-storage

An ongoing reseach project, focusing on perfomance analysis of various load
balancing techniques in a P2P system.

# Simulation

TODO: redo docker compose and update the docs

# Development

For the demo, we will use Docker and docker-compose to automate the code
compilation and initialization of the required network topology processes.

For development however, execute the binaries:

```sh
go run ./cmd/<binary>
```

You will likely not have to build the binary before execution, but if needed:

```sh
go build ./cmd/<binary>
```

## Project Directory Structure

| Type             | Path                         |
| :--------------- | :--------------------------- |
| Executables      | `./cmd/<binary>/main.go`     |
| Private packages | `./internal/<binary-name>/`  |
| Shared packages  | `./internal/<package-name>/` |

## Tests

For any tests you write, put it in the same directory as your code with the
same filename, with a `_test` suffix before file extension. For ie, the tests
for the code in `./foo/bar.go` should be in `./foo/bar_test.go`.

Go build tool includes testing with `go test`

```sh
# to run all tests
go test -v ./...

# or give it a path to a package you want to run the test
go test -v ./foo/
```

For more information, run:

```sh
go help test
```

## Code Validation

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
  - utilizing SQLite.
  - no communication needed between data nodes.
  - all nodes have the same data.
  - directly connect to the users for the data transfer.
  - encryption added, but set with a flag.
  - updating states to LB.
  - can simulate server work with `sleep`.
  - node leaving, needs to communicate with the LB.
