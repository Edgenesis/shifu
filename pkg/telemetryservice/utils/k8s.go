package utils

import (
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"os"
)

var clientSet *kubernetes.Clientset
var ns string

func initClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	ns = os.Getenv("EDGEDEVICE_NAMESPACE")
	return nil
}

func GetPasswordFromSecret(name string) (string, error) {
	if clientSet == nil {
		err := initClient()
		if err != nil {
			klog.Errorf("Can't init k8s client: %v", err)
			return "", err
		}
	}
	secret, err := clientSet.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	pwd, exist := secret.Data["password"]
	if !exist {
		return "", errors.New("the 'password' field not found in telemetry secret")
	}
	return string(pwd), nil
}
