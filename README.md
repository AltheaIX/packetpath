# Packet Path

## Status
TODO:
- 6-byte header carrying the original IP + Port
- Multi-connection support
- Revalidate with real game traffic
- Dynamically detect IP based on Program Name

## Measurements

Tested on AION2's connection to Asia Server (Indonesia -> Aion Asia Server), 3-minute captures, `tcp.analysis.ack_rtt` (TCP-level RTT).

| Metric      | Direct   | Via relay |
|-------------|----------|-----------|
| Average RTT | 36.33 ms | 35.94 ms  |
| Std Dev     | 20.99 ms | 20.58 ms  |
| P95         | 68.34 ms | 67.41 ms  |
| P99         | 75.67 ms | 75.61 ms  |

Result: no meaningful jitter reduction for this endpoint. The direct path to Cloudflare's edge is already optimal. The ~20 ms std dev is dominated by last-mile (PC→ISP) jitter, which a relay cannot fix.

```shell
ping N.N.N.N

Pinging N.N.N.N with 32 bytes of data:
Reply from N.N.N.N: bytes=32 time=29ms TTL=247
Reply from N.N.N.N: bytes=32 time=26ms TTL=247
Reply from N.N.N.N: bytes=32 time=27ms TTL=247
Reply from N.N.N.N: bytes=32 time=26ms TTL=247
```

Ping was already very optimal.

Full dataset and methodology: [`docs/measurements/jitter.md`](docs/measurements/jitter.md)

## Validation tools
- Wireshark for packet tracing
- Netcat for sending packet simulation

## Running
- Client `go run ./cmd/client` (Windows, make sure run as Administrator)
- Relay Server (VPS) `go run ./cmd/server` (Linux, hard-coded to listen on `:2323` port)

Test With:
```shell
echo "HELLO SERVER" | ncat -w 3 <game-server-ip> <game-server-port>
```

Expected:
```shell
VPS logs
[date : time] connection from <your-ip>
[date : time] read N bytes: hex
```