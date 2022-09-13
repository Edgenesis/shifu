package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	URL_EXTERNAL_IP                  = "http://cip.cc"
	URL_IP_LINE                      = "<pre>IP"
	URL_SHIFU_TELEMETRY              = "https://telemetry.shifu.run/shifu-telemetry/"
	URL_DEFAULT_PUBLIC_IP            = "0.0.0.0"
	NAMESPACE_SHIFU_CRD_SYSTEM       = "shifu-crd-system"
	DEPLOYMENTS_SHIFU_CRD_CONTROLLER = "shifu-crd-controller-manager"
	TASK_RUN_DEMO_KIND               = "run_shifu_release"
	STEP_CONTROLLER_READY            = "shifu_controller_ready"
	SOURCE_SHIFU_CONTROLLER          = "shifu_controller"
	HTTP_CONTENT_TYPE_JSON           = "application/json"
)

func GetPublicIPAddr(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting public IP")
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error getting response of IP query")
			return "", err
		}

		responseText := string(bodyBytes)
		scanner := bufio.NewScanner(strings.NewReader(responseText))
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), URL_IP_LINE) {
				ipString := strings.Split(scanner.Text(), ": ")
				return ipString[len(ipString)-1], nil
			}
		}

	}
	return "", errors.New("Did not find IP in return query")
}

func SendTelemetry(telemetry interface{}) error {
	postBodyJson, err := json.Marshal(telemetry)
	if err != nil {
		log.Printf("Error sending telemetry")
		return err
	}

	resp, err := http.Post(URL_SHIFU_TELEMETRY, HTTP_CONTENT_TYPE_JSON, bytes.NewBuffer(postBodyJson))
	if err != nil {
		log.Println("error posting telemetry, errors: ", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}
