package interceptor

import (
	"fmt"

	"github.com/one-api/godivert/types"
)

func intercept(sockData *types.DataSocket, programName string) {
	currProgName := GetProcessName(sockData.ProcessId)

	if currProgName != programName {
		return
	}

	remoteAddr := fmt.Sprintf(
		"%d.%d.%d.%d",
		byte(sockData.RemoteAddr[0]>>24),
		byte(sockData.RemoteAddr[0]>>16),
		byte(sockData.RemoteAddr[0]>>8),
		byte(sockData.RemoteAddr[0]),
	)
	remotePort := sockData.RemotePort

	localAddr := fmt.Sprintf(
		"%d.%d.%d.%d",
		byte(sockData.LocalAddr[0]>>24),
		byte(sockData.LocalAddr[0]>>16),
		byte(sockData.LocalAddr[0]>>8),
		byte(sockData.LocalAddr[0]),
	)
	localPort := sockData.LocalPort

	if remotePort == 443 || remotePort == 80 {
		return
	}

	if isLoopbackIPv4(localAddr) || isLoopbackIPv4(remoteAddr) {
		return
	}

	key := MakeSocketProgramKey(localAddr, localPort, remoteAddr, remotePort)
	SocketProgramMap.Store(key, currProgName)
}
