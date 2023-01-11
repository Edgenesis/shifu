package rtspRecord

import (
	"os/exec"
)

type Request interface {
	RegisterRequest | UnregisterRequest | UpdateRequest
}

type RegisterRequest struct {
	DeviceName    string `json:"deviceName"`
	SecretName    string `json:"secretName"`
	ServerAddress string `json:"serverAddress"`
	Recoding      bool   `json:"recoding"`
}

type UnregisterRequest struct {
	DeviceName string `json:"deviceName"`
}

type UpdateRequest struct {
	DeviceName string `json:"deviceName"`
	Record     bool   `json:"record"`
}

type Device struct {
	in      string
	cmd     *exec.Cmd
	running bool
	clip    int
}
