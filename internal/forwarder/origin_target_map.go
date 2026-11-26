package forwarder

import "sync"

var OriginTargetMap sync.Map

type OriginTargetKey struct {
	Ip   []byte
	Port uint16
}

func MakeOriginTargetKey(originalIp []byte, originalPort uint16) OriginTargetKey {
	return OriginTargetKey{
		Ip:   originalIp,
		Port: originalPort,
	}
}
