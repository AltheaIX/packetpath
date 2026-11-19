# Packet Path

## Status
TODO:
- Relay Server forwards to the original destination
- 6-byte header carrying the original IP + Port
- Multi-connection support
- Revalidate with real game traffic
- Dynamically detect IP based on Program Name

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