# 0001: Use GoDivert on Client-Side

## Status
Accepted

## Context
The goal is to forward the game's TCP traffic to VPS relay. The client (Windows PC) needs to intercept outbound and inbound TCP packets, rewrite the destination, and reinject them without modifying the game files.

## Decision
We will use WinDivert to intercept outbound and inbound TCP Packets on the client-side. We choose to use `one-api/godivert` library. Packets will be diverted at the network layer then modified and then reinjected.

## Consequences
Pros:
- WinDivert drivers directly bundled into Go binary (native)
- No virtual NIC required

Cons:
- Requires Administrator Privileges to run.