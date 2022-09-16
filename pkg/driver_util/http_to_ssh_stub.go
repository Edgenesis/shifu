package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

// Get the required configuration from the environment variables
var (
	privateSSHKeyFile    = os.Getenv("EDGEDEVICE_DRIVER_SSH_KEY_PATH")
	driverHTTPPort       = os.Getenv("EDGEDEVICE_DRIVER_HTTP_PORT")
	sshExecTimeoutSecond = os.Getenv("EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND")
	sshUser              = os.Getenv("EDGEDEVICE_DRIVER_SSH_USER")
)

func init() {
	if privateSSHKeyFile == "" {
		klog.Fatalf("SSH Keyfile needs to be specified")
	}

	if driverHTTPPort == "" {
		driverHTTPPort = "11112"
		klog.Infof("No HTTP Port specified for driver, default to %v", driverHTTPPort)
	}

	if sshExecTimeoutSecond == "" {
		sshExecTimeoutSecond = "5"
		klog.Infof("No SSH exec timeout specified for driver, default to %v seconds", sshExecTimeoutSecond)
	}

	if sshUser == "" {
		sshUser = "root"
		klog.Infof("No SSH user specified for driver, default to %v", sshUser)
	}
}

func main() {
	key, err := ioutil.ReadFile(privateSSHKeyFile)
	if err != nil {
		klog.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		klog.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout:         time.Minute,
	}

	sshClient, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		klog.Fatalf("unable to connect: %v", err)
	}
	defer sshClient.Close()
	klog.Infof("Driver SSH established")

	sshListener, err := sshClient.Listen("tcp", "localhost:"+driverHTTPPort)
	if err != nil {
		klog.Fatalf("unable to register tcp forward: %v", err)
	}
	defer sshListener.Close()
	klog.Infof("Driver HTTP listener established")

	err = http.Serve(sshListener, httpCmdlinePostHandler(sshClient))
	if err != nil {
		klog.Errorf("cannot start server, error: %v", err)
	}
}

// Create a session reply for the incoming connection, obtain the connection body information,
// process it, hand it over to the shell for processing,return both result and status code based
// on shell execution result
func httpCmdlinePostHandler(sshConnection *ssh.Client) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		session, err := sshConnection.NewSession()
		if err != nil {
			klog.Fatalf("Failed to create session: %v", err)
		}

		defer session.Close()
		httpCommand, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		cmdString := "timeout " + sshExecTimeoutSecond + " " + string(httpCommand)
		klog.Infof("running command: %v", cmdString)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		session.Stdout = &stdout
		session.Stderr = &stderr
		if err := session.Run(cmdString); err != nil {
			klog.Errorf("Failed to run cmd: %v\n stderr: %v \n stdout: %v", cmdString, stderr.String(), stdout.String())
			resp.WriteHeader(http.StatusBadRequest)
			_, _ = resp.Write(append(stderr.Bytes(), stdout.Bytes()...))
			return
		}

		klog.Infof("cmd: %v success", cmdString)
		resp.WriteHeader(http.StatusOK)
		_, _ = resp.Write(stdout.Bytes())
	}
}
