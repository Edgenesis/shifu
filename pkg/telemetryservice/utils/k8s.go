package utils

import (
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

var clientSet *kubernetes.Clientset
var ns string

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	nsb, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err.Error())
	}
	ns = string(nsb)
}

func GetPasswordFromSecret(name string) (string, error) {
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
