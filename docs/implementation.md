# Implementation

TODO: table of content

## Port Forwarding Process

1. User opens a connection to the LB.
2. LB allocates a 7 byte address buffer, the first byte represent a message and
   has the byte value of `network.UserNodeJoin`, the remaining 6 bytes
   represent the ipv4 address (4 bytes), and port number (2 bytes).
3. LB selects a data node to distribute the request to, depends on the
   algorithms selected for the simulation, then send this address buffer to
   the selected node.
4. The data node receives this (7 byte) message, opens a TCP listener server
   for the requesting user, changes the message type to `network.DataNodePort`
   then appends this address to the message buffer, making it 13 byte long,
   then sends this back to the LB.
5. LB receives a `network.DataNodePort` message `buf`, parse the user address
   from `buf[1:7]`, then forward the entire (13 bytes) message to the user of
   the address.
6. User receives this message and dials data node, with the adddress parsed
   from `buf[7:13]`, completes the port forwarding process.
