package deviceshifusocket

// RequestBody Socket Request Body
type RequestBody struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

// ReturnBody Socket Reply Body
type ReturnBody struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}
