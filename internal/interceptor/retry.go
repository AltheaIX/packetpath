package interceptor

import "time"

func ResolveProgramName(key SocketProgramKey) string {
	const (
		maxAttempt = 5
		retryDelay = 200 * time.Nanosecond
	)

	var attempt = 0
	for attempt = range maxAttempt {
		if val, ok := SocketProgramMap.Load(key); ok {
			if name, ok := val.(string); ok && name != "" {
				return name
			}
		}
		if attempt < maxAttempt-1 {
			time.Sleep(retryDelay)
		}
	}

	return "unknown"
}
