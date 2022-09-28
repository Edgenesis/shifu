package utils

import (
	"reflect"
	"testing"
)

func TestParseHTTPGetParams(t *testing.T) {
	url := "http://www.example.com?a=1&b=22&c=3&a=5&d="
	currentParam := map[string]string{
		"a": "1",
		"b": "22",
		"c": "3",
		"d": "",
	}

	params, err := ParseHTTPGetParams(url)
	if err != nil {
		t.Errorf("Error when ParseHTTPGetParams, %v", err)
	}
	if !reflect.DeepEqual(currentParam, params) {
		t.Errorf("Not math current Params,out: %v,current:%v", params, currentParam)
	}
}
