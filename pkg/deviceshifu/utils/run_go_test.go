package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunGo(t *testing.T) {
	rawData := "{\"statusCode\":\"200\",\"message\":\"success\",\"entity\":[{\"deviceId\":\"20990922009\",\"datatime\":\"2022-06-30 07:55:51\",\"eUnit\":\"℃\",\"eValue\":\"37\",\"eKey\":\"e3\",\"eName\":\"atmosphere temperature\",\"eNum\":\"101\"},{\"deviceId\":\"20990922009\",\"datatime\":\"2022-06-30 07:55:51\",\"eUnit\":\"%RH\",\"eValue\":\"88\",\"eKey\":\"e4\",\"eName\":\"atmosphere humidity\",\"eNum\":\"102\"}]}"

	res := Run("./humidity", rawData)

	// need a runnable binary exist under go_custom_handler
	assert.Equal(t, "", res)
}
