package minio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifuhttp"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/pkg/telemetryservice/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/http"
	"time"
)

func BindMinIOServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	request := v1alpha1.TelemetryRequest{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct")
		http.Error(w, "Unexpected end of JSON input", http.StatusBadRequest)
		return
	}
	if request.MinIOSetting.Bucket == nil || request.MinIOSetting.EndPoint == nil || request.MinIOSetting.FileExtension == nil {
		logger.Errorf("Bucket or EndPoint or FileExtension cant be nil")
		http.Error(w, "Bucket or EndPoint or FileExtension cant be nil", http.StatusBadRequest)
		return
	}
	// Read MinIo APIId & APIKey
	injectSecret(request.MinIOSetting)
	if request.MinIOSetting.APIId == nil || request.MinIOSetting.APIKey == nil {
		logger.Errorf("Fail to get APIId or APIKey")
		http.Error(w, "Fail to get APIId or APIKey", http.StatusBadRequest)
		return
	}

	// Create MinIO Client
	client, err := minio.New(*request.MinIOSetting.EndPoint, &minio.Options{
		Creds: credentials.NewStaticV4(*request.MinIOSetting.APIId, *request.MinIOSetting.APIKey, ""),
	})
	if err != nil {
		logger.Errorf("Fail to create MinIO client, error:" + err.Error())
		http.Error(w, "Fail to create MinIO client", http.StatusBadRequest)
		return
	}
	// Get device name, build file path ([device_name]/[time].[fileExtension])
	if deviceName, ok := r.Header[deviceshifuhttp.DeviceNameHeaderField]; ok && len(deviceName) > 0 {
		fileName := fmt.Sprintf("%v/%v.%v", deviceName[0], time.Now().Format(time.RFC3339), *request.MinIOSetting.FileExtension)
		// Upload file to MinIO
		err = uploadObject(client, *request.MinIOSetting.Bucket, fileName, request.RawData)
		if err != nil {
			logger.Errorf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	} else {
		logger.Errorf("Fail to get device name from header")
		http.Error(w, "Fail to get device name from header", http.StatusBadRequest)
	}
}

func injectSecret(setting *v1alpha1.MinIOSetting) {
	if setting == nil {
		logger.Warn("Empty MinIO service setting")
		return
	}
	if setting.Secret == nil {
		logger.Warn("Empty MinIO secret setting")
		return
	}
	secret, err := utils.GetSecret(*setting.Secret)
	if err != nil {
		logger.Errorf("Fail to get secret of %v, error: %v", *setting.Secret, err)
		return
	}
	// Load APIId & APIKey from secret
	if id, exist := secret[deviceshifubase.UsernameSecretField]; exist {
		setting.APIId = &id
	} else {
		logger.Errorf("Fail to get APIId from secret")
		return
	}
	if key, exist := secret[deviceshifubase.PasswordSecretField]; exist {
		setting.APIKey = &key
	} else {
		logger.Errorf("Fail to get APIKey from secret")
		return
	}

	logger.Infof("MinIo loaded APIId & APIKey from secret")
}

func uploadObject(client *minio.Client, bucket string, fileName string, content []byte) error {
	// Read file's content
	reader := bytes.NewReader(content)
	// Upload file
	_, err := client.PutObject(context.Background(),
		bucket, fileName, reader, reader.Size(),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		logger.Error("Upload object error:" + err.Error())
		return errors.New("upload object fail")
	}
	logger.Infof("Upload file success, fileName:%v", fileName)
	return nil
}
