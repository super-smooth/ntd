package tailscale

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Host represents a Tailscale host
type Host struct {
	Hostname  string
	IP        string
	OS        string
	Status    string
	IsCurrent bool
	IsNixOS   bool // Will be determined by SSH check
}

// GetHosts retrieves all hosts from the Tailscale network
func GetHosts() ([]Host, error) {
	cmd := exec.Command("tailscale", "status")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("tailscale status failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run tailscale status: %w", err)
	}

	return parseStatus(string(output))
}

// GetCurrentHostname returns the current machine's hostname
func GetCurrentHostname() (string, error) {
	cmd := exec.Command("tailscale", "status", "--self")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to system hostname
		hostname, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("failed to get hostname: %w", err)
		}
		return hostname, nil
	}

	// Parse the first line to get current hostname
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			return fields[1], nil
		}
	}

	// Fallback to system hostname
	return os.Hostname()
}

// FilterLinuxHosts returns only Linux hosts (potential NixOS hosts)
func FilterLinuxHosts(hosts []Host) []Host {
	var linuxHosts []Host
	for _, host := range hosts {
		// Check if OS contains "linux" (case insensitive)
		if strings.Contains(strings.ToLower(host.OS), "linux") {
			linuxHosts = append(linuxHosts, host)
		}
	}
	return linuxHosts
}

// parseStatus parses the output of `tailscale status`
func parseStatus(output string) ([]Host, error) {
	var hosts []Host
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Get current hostname to mark it in the list
	currentHostname, _ := GetCurrentHostname()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Format: 100.x.x.x hostname user@ os status
		host := Host{
			IP:        fields[0],
			Hostname:  fields[1],
			OS:        fields[3],
			IsCurrent: fields[1] == currentHostname,
		}

		if len(fields) > 4 {
			host.Status = fields[4]
		}

		hosts = append(hosts, host)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse tailscale status: %w", err)
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no Tailscale hosts found")
	}

	return hosts, nil
}
