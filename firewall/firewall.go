package firewall

import (
	"os/exec"
)

// Executor defines the interface for executing shell commands.
type Executor interface {
	Execute(name string, arg ...string) ([]byte, error)
}

// RealExecutor is the production implementation that runs actual commands.
type RealExecutor struct{}

func (e *RealExecutor) Execute(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput()
}

// Firewall handles the interaction with iptables and tc.
type Firewall struct {
	exec          Executor
	iface         string
	wanIface      string
	qdiscInit     bool
	DownloadSpeed string
	UploadSpeed   string
}

func New(iface string, wanIface string, exec Executor) *Firewall {
	if exec == nil {
		exec = &RealExecutor{}
	}
	return &Firewall{
		exec:          exec,
		iface:         iface,
		wanIface:      wanIface,
		DownloadSpeed: "0",
		UploadSpeed:   "0",
	}
}

// SetExecutor replaces the command executor (useful for testing).
func (f *Firewall) SetExecutor(exec Executor) {
	f.exec = exec
}
