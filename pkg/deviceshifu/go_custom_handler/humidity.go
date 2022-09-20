package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strconv"

	"github.com/buger/jsonparser"
	"k8s.io/klog/v2"
)

// explicity declare the data structure make the parser result much straitforward
type ProcessedData struct {
	Code      int64  `json:"code"`
	Name      string `json:"name"`
	Val       int64  `json:"val"`
	Unit      string `json:"unit"`
	Exception string `json:"exception"`
}

const TEMPERATURE_MEASUREMENT = "atmosphere temperature"
const HUMIDITY_MEASUREMENT = "atmosphere humidity"

func checkRegularMeasurementException(measurementName string, measurementValue int64) string {
	exceptionMsg := ""
	if measurementName == TEMPERATURE_MEASUREMENT {
		if measurementValue > 35 {
			exceptionMsg = "temperature is too high"
		}
	} else if measurementName == HUMIDITY_MEASUREMENT {
		if measurementValue > 60 {
			exceptionMsg = "humidity is too high"
		}
	}
	return exceptionMsg
}

// the implementation need to know the json reading path from the raw data
func humidity(rawData string) string {
	//TODO: collect and log error if any
	var newData []ProcessedData
	jsonparser.ArrayEach([]byte(rawData), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		eName, _ := jsonparser.GetString(value, "eName")
		eValueStr, _ := jsonparser.GetString(value, "eValue")
		eCodeStr, _ := jsonparser.GetString(value, "deviceId")
		eUnit, _ := jsonparser.GetString(value, "eUnit")
		eCode, _ := strconv.ParseInt(eCodeStr, 10, 64)
		eValue, _ := strconv.ParseInt(eValueStr, 10, 64)
		newEntry := ProcessedData{
			Code:      eCode,
			Name:      eName,
			Val:       eValue,
			Unit:      eUnit,
			Exception: checkRegularMeasurementException(eName, eValue),
		}
		newData = append(newData, newEntry)
	}, "entity")

	res, _ := json.Marshal(newData)

	return string(res)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		klog.Error("go custom handler need at least one parameter to run")
		return
	}
	w := bufio.NewWriter(os.Stdout)
	io.Writer.Write(w, []byte(humidity(args[1])))
	w.Flush()
}
