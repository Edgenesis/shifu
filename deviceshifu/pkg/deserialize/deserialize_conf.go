package deserialize

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"knative.dev/pkg/configmap"
	"log"
	"strconv"
)

type ResponseMsgTypeInDictionaries struct {
	ResponseMsgType string `yaml:"responseMsgType"`
}

type FormattedStringDictionaries struct {
	DictionaryDilimiter string   `yaml:"dictionaryDilimiter"`
	Dictionary          []string `yaml:"dictionary"`
}

type DecodeConfigStr struct {
	Enabled            string                 `yaml:"enabled"`
	DecodeDictionaries map[string]interface{} `yaml:"decodeDictionaries"`
}

const (
	CM_DECODECONFIG_STR    = "decoding"
	DECODECONFIG_FILE_PATH = "/etc/edgedevice/config/"
)

var decodeConfig DecodeConfigStr

func init() {
	var err error
	cfg, err := configmap.Load(DECODECONFIG_FILE_PATH)
	log.Println(cfg)
	if err != nil {
		log.Println(err)
	}
	if decodeConfs, ok := cfg[CM_DECODECONFIG_STR]; ok {
		log.Println(decodeConfs)
		err = yaml.Unmarshal([]byte(decodeConfs), &decodeConfig)

		if err != nil {
			log.Println(err)
		}
	}
	if decodeConfig.DecodeDictionaries == nil {
		log.Println("Configuration file is empty")
		enabled = false
	}

	enabled, _ = strconv.ParseBool(decodeConfig.Enabled)
	log.Println(decodeConfig)

	if enabled {
		for k := range decodeConfig.DecodeDictionaries {
			InstructionsTobeDeserialize = append(InstructionsTobeDeserialize, k)
		}
	}

	log.Println(InstructionsTobeDeserialize)
}

func GetRespontsMsgType(key string) (string, error) {
	var ResponseMsgType ResponseMsgTypeInDictionaries
	configToByte, err := json.Marshal(decodeConfig.DecodeDictionaries[key])
	if err != nil {
		log.Println(err)
		return "", err
	}

	err = yaml.Unmarshal(configToByte, &ResponseMsgType)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return ResponseMsgType.ResponseMsgType, nil
}

func GetformattedConf(key string) (FormattedStringDictionaries, error) {
	var FormattedStringConfig FormattedStringDictionaries
	ConfigToByte, err := json.Marshal(decodeConfig.DecodeDictionaries[key])
	if err != nil {
		log.Println(err)
		return FormattedStringDictionaries{}, err
	}

	err = yaml.Unmarshal(ConfigToByte, &FormattedStringConfig)
	if err != nil {
		log.Println(err)
		return FormattedStringDictionaries{}, err
	}

	return FormattedStringConfig, nil
}
