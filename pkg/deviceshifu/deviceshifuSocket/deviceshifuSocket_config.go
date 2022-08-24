package deviceshifuSocket

type DeviceShifuSocketRequestBody struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type DeviceShifuSocketReturnBody struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}
