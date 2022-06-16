package deserialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ObjectDeserialize struct {
	responseMsgType             string
	instructions                string
	ResponseMsg                 string
	jsonResponseMsg             []byte
	FormattedStringDictionaries FormattedStringDictionaries
}

type ResponseData struct {
	RawData    string
	FormatData map[string]string
}

const (
	responseMsgTypeFormattedString = "formattedString"
	responseMsgTypeJson            = "json"
	responseMsgTypeString          = "string"
)

var (
	InstructionsTobeDeserialize []string
	enabled                     bool
)

func NewReader(r *url.URL, resp *http.Response) (*ObjectDeserialize, error) {
	var ResponseMsg = resp.Body
	objectDeserialize := new(ObjectDeserialize)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(ResponseMsg)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	objectDeserialize.ResponseMsg = buf.String()
	if !enabled || resp.StatusCode >= http.StatusBadRequest {
		return objectDeserialize,
			fmt.Errorf("configuration or request error")
	}

	for _, v := range InstructionsTobeDeserialize {
		deserializeOrNot := strings.Contains(r.String(), v)
		if deserializeOrNot {
			objectDeserialize.instructions = v
			break
		}
	}

	if objectDeserialize.instructions == "" {
		log.Println("Unconfigured instruction",
			objectDeserialize.ResponseMsg)
		return objectDeserialize,
			fmt.Errorf("This instruction does not exist in the dictionary")
	}

	objectDeserialize.responseMsgType, _ =
		GetRespontsMsgType(objectDeserialize.instructions)
	switch objectDeserialize.responseMsgType {
	case responseMsgTypeFormattedString:
		objectDeserialize.FormattedStringDictionaries,
			err = GetformattedConf(objectDeserialize.instructions)
		if err != nil {
			log.Println(err)
			return objectDeserialize, err
		}
	case responseMsgTypeJson:
		objectDeserialize.jsonResponseMsg = buf.Bytes()
	}

	return objectDeserialize, err
}

func (DeserializeObject *ObjectDeserialize) Deserialize() (io.Reader, error) {

	switch DeserializeObject.responseMsgType {
	case responseMsgTypeFormattedString:
		return FormattedStringDeserialize(DeserializeObject)
	case responseMsgTypeJson:
		return JsonDeserialize(DeserializeObject)
	case responseMsgTypeString:
		return StringDeserialize(DeserializeObject)
	}

	return strings.NewReader(DeserializeObject.ResponseMsg),
		fmt.Errorf("there is no corresponding configuration" +
			" in the configuration file,the original format is returned")
}

func FormattedStringDeserialize(DeserializeObject *ObjectDeserialize) (io.Reader, error) {
	var responseData ResponseData
	responseData.FormatData = make(map[string]string)

	var dictionary = DeserializeObject.
		FormattedStringDictionaries.Dictionary
	var formattedStringDilimiter = DeserializeObject.
		FormattedStringDictionaries.DictionaryDilimiter
	dataArray := strings.Split(DeserializeObject.ResponseMsg,
		formattedStringDilimiter)
	if strings.Count(DeserializeObject.ResponseMsg,
		formattedStringDilimiter) == len(dictionary)-1 {
		for k, v := range dataArray {
			if dictionary[k] != "" {
				responseData.RawData = DeserializeObject.ResponseMsg
				responseData.FormatData[dictionary[k]] = v
			}
		}

		log.Println("The original string of the device response is",
			DeserializeObject.ResponseMsg)

		FormattedStringJson, err := json.Marshal(responseData)
		return bytes.NewReader(FormattedStringJson), err
	}

	log.Println("String does not correspond to dictionary")
	return strings.NewReader(DeserializeObject.ResponseMsg),
		fmt.Errorf("String does not correspond to dictionary")
}

func JsonDeserialize(DeserializeObject *ObjectDeserialize) (io.Reader, error) {
	var result map[string]interface{}
	err := json.Unmarshal(DeserializeObject.jsonResponseMsg, &result)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("The device response json :", result)
	return bytes.NewReader(DeserializeObject.jsonResponseMsg), nil
}

func StringDeserialize(DeserializeObject *ObjectDeserialize) (io.Reader, error) {
	log.Print("The device response string :", DeserializeObject.ResponseMsg)

	return strings.NewReader(DeserializeObject.ResponseMsg), nil
}
