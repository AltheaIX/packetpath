# 0002: Use Custom Header

## Status
Accepted

## Context
The Relay Server needs to know where the forwarded packet need to be sent. For that matters, original destination IP and Port need to be provided for each connection to fulfill the multi-connection needs.

Alternative considered:
- Hardcode the destination IP (not dynamic and breaks when game switches servers)
- Port mapping (one VPS port per destination, not dynamic)

## Decision
We will send 6-byte custom header (4 byte IP and 2 byte Port) once per connection, carrying the original destination.


## Consequences
Pros:
- Multi-Connection supports to different destination
- Stateless

Cons:
- 6-byte overhead every a new connection made
- The relay must read and strip the header before forwarding