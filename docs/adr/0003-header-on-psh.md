# 0003: Header on PSH

## Context
Header was originally placed in SYN packet. However, net.Listener.Accept() on Go doesn't catch connection handshake (SYN, SYN-ACK, ACK). So the header cannot be read from the SYN on the server side.

Alternative considered:
- Packet Interceptor on Server to catch data on SYN (complex)

## Decision
We will send the custom header on the first PSH instead of SYN. The VPS accepts the connection, reads the 6-byte header from the stream (once), then dials the original destination.

## Consequences
Pros:
- No interceptor needed on the server side

Cons:
- The client must track first PSH per connection
- The client must shift sequence number (+6) on outbound and shift ack number (-6) on inbound

Neutral:
- Dial to the original destination happens after first PSH and not after SYN.