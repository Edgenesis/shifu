package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
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
		log.Fatalf("SSH Keyfile needs to be specified")
	}

	if driverHTTPPort == "" {
		driverHTTPPort = "11112"
		log.Printf("No HTTP Port specified for driver, default to %v", driverHTTPPort)
	}

	if sshExecTimeoutSecond == "" {
		sshExecTimeoutSecond = "5"
		log.Printf("No SSH exec timeout specified for driver, default to %v seconds", sshExecTimeoutSecond)
	}

	if sshUser == "" {
		sshUser = "root"
		log.Printf("No SSH user specified for driver, default to %v", sshUser)
	}
}

func main() {
	key, err := ioutil.ReadFile(privateSSHKeyFile)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
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
		log.Fatal("unable to connect: ", err)
	}
	defer sshClient.Close()
	log.Println("Driver SSH established")

	ssh_listener, err := sshClient.Listen("tcp", "localhost:"+driverHTTPPort)
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}
	defer ssh_listener.Close()
	log.Println("Driver HTTP listener established")

	http.Serve(ssh_listener, httpCmdlinePostHandler(sshClient))
}

// Create a session reply for the incoming connection, obtain the connection body information,
// process it, hand it over to the shell for processing,return both result and status code based
// on shell execution result
func httpCmdlinePostHandler(sshConnection *ssh.Client) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		session, err := sshConnection.NewSession()
		if err != nil {
			log.Fatal("Failed to create session: ", err)
		}

		defer session.Close()
		httpCommand, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		cmdString := "timeout " + sshExecTimeoutSecond + " " + string(httpCommand)
		log.Printf("running command: %v\n", cmdString)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		session.Stdout = &stdout
		session.Stderr = &stderr
		if err := session.Run(cmdString); err != nil {
			log.Printf("Failed to run cmd: %v\n stderr: %v \n stdout: %v", cmdString, stderr.String(), stdout.String())
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write(append(stderr.Bytes(), stdout.Bytes()...))
			return
		}

		log.Printf("cmd: %v success", cmdString)
		resp.WriteHeader(http.StatusOK)
		resp.Write(stdout.Bytes())
	}
}
