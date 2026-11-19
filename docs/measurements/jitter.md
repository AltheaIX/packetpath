# Jitter Measurement

Date: 11/20/2026 ~01:00 AM

Client: Windows 11, RZ608 Wi-Fi 6E, Indonesia

Relay: Singapore VPS (Digital Ocean Provider)

Game State: Idle, no player around, no combat, no zone changes

## Methodology
### Capturing Data
Data was captured with 2 Wireshark Filter for Relay and Non-Relay (Direct) connection.
```shell
tcp && (ip.addr == gameIP) - for Direct
tcp && (ip.addr == relayIP) - for Relay 
```
Data was captured for 3-minutes long from the connection creation (SYN, SYN-ACK) in a quiet map (almost no player situation) and idling.

### Processing Data
Data was processed with Tshark with filtering to tcp.port,
```shell
Direct Data Command:
tshark -r aion-no-relay_3min.pcapng -Y "tcp.port == 13328 && tcp.analysis.ack_rtt" -T fields -e tcp.analysis.ack_rtt > direct_rtt.txt

Relay Data Command:
tshark -r aion-with-relay_3min.pcapng -Y "tcp.port == 2323 && tcp.analysis.ack_rtt" -T fields -e tcp.analysis.ack_rtt > relay_rtt.txt
```

## Raw Data
Direct Data: [Here](direct_rtt.txt), Samples: 6931 data.

Relay Data: [Here](relay_rtt.txt), Samples: 7061 data.

Original .pcapng captures weren't provided because they contain full packet headers including IP addresses and payload bytes.

## Analysis
RTT values (seconds) were converted to milliseconds and analyzed for average, median, standard deviation, and percentiles (P95, P99). The metrics below are computed over the full 3-minute capture.

| Metric         | Direct    | Via relay | Difference       |
|----------------|-----------|-----------|------------------|
| Samples (ACKs) | 6931      | 7061      | ~                |
| Average RTT    | 36.33 ms  | 35.94 ms  | −0.39 ms (−1.1%) |
| Median (P50)   | 36.29 ms  | 35.73 ms  | −0.56 ms (−1.5%) |
| Std Dev        | 20.99 ms  | 20.58 ms  | −0.41 ms (−2.0%) |
| P95            | 68.34 ms  | 67.41 ms  | −0.93 ms (−1.4%) |
| P99            | 75.67 ms  | 75.61 ms  | −0.06 ms (−0.1%) |
| Max            | 307.39 ms | 268.23 ms | −39.16 ms        |

### Conclusion
No meaningful jitter reduction because the direct-path already in a good route.