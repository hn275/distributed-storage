# distributed-storage

An ongoing reseach project, focusing on perfomance analysis of various load
balancing techniques in a P2P system.

## Simulation

```sh
docker compose up
```

## Requirements

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
