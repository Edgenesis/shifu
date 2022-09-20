package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumidity(t *testing.T) {

	testJson := "{\"statusCode\":\"200\",\"message\":\"success\",\"entity\":[{\"deviceId\":\"20990922009\",\"datatime\":\"2022-06-30 07:55:51\",\"eUnit\":\"℃\",\"eValue\":\"37\",\"eKey\":\"e3\",\"eName\":\"atmosphere temperature\",\"eNum\":\"101\"},{\"deviceId\":\"20990922009\",\"datatime\":\"2022-06-30 07:55:51\",\"eUnit\":\"%RH\",\"eValue\":\"88\",\"eKey\":\"e4\",\"eName\":\"atmosphere humidity\",\"eNum\":\"102\"}]}"

	result := humidity(testJson)
	assert.Equal(t, result, "[{\"code\":20990922009,\"name\":\"atmosphere temperature\",\"val\":37,\"unit\":\"℃\",\"exception\":\"temperature is too high\"},{\"code\":20990922009,\"name\":\"atmosphere humidity\",\"val\":88,\"unit\":\"%RH\",\"exception\":\"humidity is too high\"}]")
}
