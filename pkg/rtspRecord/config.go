package rtspRecord

import "os/exec"

type Request interface {
	RegisterRequest | UnregisterRequest | UpdateRequest
}

type RegisterRequest struct {
	DeviceName    string `json:"deviceName"`
	SecretName    string `json:"secretName"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ServerAddress string `json:"serverAddress"`
	Recoding      bool   `json:"recoding"`
	OutputPath    string `json:"outputPath"`
}

type UnregisterRequest struct {
	DeviceName string `json:"deviceName"`
}

type UpdateRequest struct {
	DeviceName string `json:"deviceName"`
	Record     bool   `json:"record"`
}

type Device struct {
	cmd     *exec.Cmd
	running bool
}
