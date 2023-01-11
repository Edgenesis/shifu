package rtspRecord

import (
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/pkg/telemetryservice/utils"
	"io"
	"net/http"
)

func trans[T Request](r *http.Request) (*T, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	request := new(T)
	err = json.Unmarshal(body, request)
	if err != nil {
		return nil, err
	}
	logger.Infof("request: %v", *request)
	return request, nil
}

func getCredential(name string) (string, string, error) {
	secret, err := utils.GetSecret(name)
	if err != nil {
		return "", "", err
	}
	password, exist := secret[deviceshifubase.PasswordSecretField]
	if !exist {
		return "", "", fmt.Errorf("the %v field not found in telemetry secret", deviceshifubase.PasswordSecretField)
	}
	username, exist := secret[deviceshifubase.UsernameSecretField]
	if !exist {
		return "", "", fmt.Errorf("the %v field not found in telemetry secret", deviceshifubase.UsernameSecretField)
	}
	return username, password, nil
}
