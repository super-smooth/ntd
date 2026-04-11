package deploy

import (
	"fmt"
)

// Deployer generates nixos-rebuild commands
type Deployer struct {
	FlakePath string
	Output    string
	Host      string
	IsLocal   bool
	NoSudo    bool
}

// NewDeployer creates a new Deployer instance
func NewDeployer(flakePath, output, host string, isLocal, noSudo bool) *Deployer {
	return &Deployer{
		FlakePath: flakePath,
		Output:    output,
		Host:      host,
		IsLocal:   isLocal,
		NoSudo:    noSudo,
	}
}

// GenerateCommand returns the nixos-rebuild command as a string
func (d *Deployer) GenerateCommand() string {
	if d.IsLocal {
		if d.NoSudo {
			return fmt.Sprintf("nixos-rebuild switch --flake %s#%s",
				d.FlakePath, d.Output)
		}
		return fmt.Sprintf("sudo nixos-rebuild switch --flake %s#%s",
			d.FlakePath, d.Output)
	}

	// Remote deployment - default to asking for sudo password
	if d.NoSudo {
		return fmt.Sprintf("nixos-rebuild switch --flake %s#%s --target-host %s",
			d.FlakePath, d.Output, d.Host)
	}
	return fmt.Sprintf("nixos-rebuild switch --flake %s#%s --target-host %s --sudo --ask-sudo-password",
		d.FlakePath, d.Output, d.Host)
}
