package interceptor

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

func Ipv4ToString(ip uint32) string {
	return fmt.Sprintf("%v.%v.%v.%v", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func GetProcessName(pid uint32) string {
	if pid == 0 {
		return "System (Idle)"
	}
	if pid == 4 {
		return "System"
	}

	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "unknown"
	}
	defer windows.CloseHandle(h)

	var size uint32 = 1024
	buf := make([]uint16, size)
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return "unknown"
	}

	path := syscall.UTF16ToString(buf[:size])
	// Extract base name
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
