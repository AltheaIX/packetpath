package interceptor

import "sync"

var (
	SocketProgramMap sync.Map
)

type SocketProgramKey struct {
	srcIP   string
	srcPort uint16
	dstIP   string
	dstPort uint16
}

func MakeSocketProgramKey(srcIp string, srcPort uint16, dstIp string, dstPort uint16) SocketProgramKey {
	return SocketProgramKey{
		srcIP:   srcIp,
		srcPort: srcPort,
		dstIP:   dstIp,
		dstPort: dstPort,
	}
}
