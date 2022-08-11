package utils

import (
	"reflect"
	"testing"
)

func TestParseAllParams(t *testing.T) {
	input := []string{
		`http://test.com`,
		`test.com?a=3`,
		`http://test.com?a=1&b=2`,
	}
	output := []map[string]string{
		nil,
		map[string]string{"a": "3"},
		map[string]string{"a": "1", "b": "2"},
	}

	for id, item := range input {
		out := ParseAllParams(item)
		if !reflect.DeepEqual(out, output[id]) {
			t.Error()
		}
	}
}
