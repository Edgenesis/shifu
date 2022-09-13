package deviceshifuSocket

type DeviceShifuSocketRequestBody struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
	Encode  string `json:"Encode"`
}

type DeviceShifuSocketReturnBody struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

const (
	MessageEncodeUTF8    = "utf8"
	MessageEncodeHEX     = "hex"
	MessageEncodeUnicode = "unicode"
)
