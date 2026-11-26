package config

import (
	"fmt"
	"strconv"
)

type Config struct {
	ProgramName string
	RelayIp     string
	RelayPort   uint16
}

func ParsePort(rawPort string) (uint16, error) {
	p, err := strconv.Atoi(rawPort)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", rawPort)
	}

	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("port out of range: %s", rawPort)
	}

	return uint16(p), nil
}
