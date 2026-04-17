package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

/*
*
IF Log file - good
No log file - ignore (it's optional)
**
*/
func setupLogFile(path string) *os.File {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	file, err := initLogFile(path)
	if err != nil {
		return nil
	}

	return file
}

func initLogFile(logFile string) (*os.File, error) {
	file, err := os.OpenFile(
		logFile,
		os.O_WRONLY|
			os.O_CREATE|
			os.O_APPEND,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("Error opening Log file: %w", err)
	}
	return file, nil
}

func redactIP(addStr string) string {
	if strings.TrimSpace(addStr) == "" {
		return ""
	}

	host, _, err := net.SplitHostPort(
		addStr,
	)
	if err != nil {
		return addStr
	}

	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return host
	}

	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return host
	}

	return fmt.Sprintf("%d.%d.%d.x", ipv4[0], ipv4[1], ipv4[2])
}
