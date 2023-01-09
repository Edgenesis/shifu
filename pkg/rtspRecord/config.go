package rtspRecord

import (
	"os/exec"
	"sync"
)

type Request interface {
	RegisterRequest | UnregisterRequest | UpdateRequest
}

type RegisterRequest struct {
	DeviceName    string `json:"deviceName"`
	SecretName    string `json:"secretName"`
	ServerAddress string `json:"serverAddress"`
	Recoding      bool   `json:"recoding"`
	OutDir        string `json:"outDir"`
}

type UnregisterRequest struct {
	DeviceName string `json:"deviceName"`
}

type UpdateRequest struct {
	DeviceName string `json:"deviceName"`
	Record     bool   `json:"record"`
}

type Device struct {
	mu      sync.Mutex
	in      string
	outDir  string
	cmd     *exec.Cmd
	running bool
	clip    int
}
