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

func main() {
	privateSSHKeyFile := os.Getenv("EDGEDEVICE_DRIVER_SSH_KEY_PATH")
	driverHTTPPort := os.Getenv("EDGEDEVICE_DRIVER_HTTP_PORT")
	sshExecTimeoutSecond := os.Getenv("EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND")
	sshUser := os.Getenv("EDGEDEVICE_DRIVER_SSH_USER")

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

	ssh_connection, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}
	defer ssh_connection.Close()
	log.Println("Driver SSH established")

	ssh_listener, err := ssh_connection.Listen("tcp", "localhost:"+driverHTTPPort)
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}
	defer ssh_listener.Close()
	log.Println("Driver HTTP listener established")

	http.Serve(ssh_listener, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		session, err := ssh_connection.NewSession()
		if err != nil {
			log.Fatal("Failed to create session: ", err)
		}

		defer session.Close()
		rb, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		cmdString := "timeout " + sshExecTimeoutSecond + " " + string(rb)
		log.Printf("running command: %v\n", cmdString)
		var stdcombined bytes.Buffer
		session.Stdout = &stdcombined
		session.Stderr = &stdcombined
		if err := session.Run(cmdString); err != nil {
			log.Printf("Failed to run cmd: %v, stderr: %v", cmdString, stdcombined.String())
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write(stdcombined.Bytes())
			return
		}

		resp.WriteHeader(http.StatusOK)
		resp.Write(stdcombined.Bytes())
	}))
}
