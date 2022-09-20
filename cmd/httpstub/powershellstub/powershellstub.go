package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

var (
	driverHTTPPort       = os.Getenv("EDGEDEVICE_DRIVER_HTTP_PORT")
	cmdExecTimeoutSecond = os.Getenv("EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND")
)

func init() {
	if driverHTTPPort == "" {
		driverHTTPPort = "11112"
		klog.Infof("No HTTP Port specified for driver, default to %v", driverHTTPPort)
	}

	if cmdExecTimeoutSecond == "" {
		cmdExecTimeoutSecond = "5"
		klog.Infof("No SSH exec timeout specified for driver, default to %v seconds", cmdExecTimeoutSecond)
	}
}

func main() {
	http.HandleFunc("/", httpCmdlinePostHandler)
	http.HandleFunc("/fileserve", httpFileServeHandler)
	http.HandleFunc("/stub_health", httpHealthHandler)
	klog.Fatal(http.ListenAndServe("0.0.0.0:"+driverHTTPPort, nil))
}

func httpCmdlinePostHandler(resp http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	for parameterName, parameterValues := range values {
		klog.Infof("paramname is: %v, value is: %v", parameterName, parameterValues[0])
	}

	timeoutSeconds, err := strconv.Atoi(cmdExecTimeoutSecond)
	if err != nil {
		klog.Infof("cannot convert cmdExecTimeoutSecond: %v to integer", cmdExecTimeoutSecond)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	timeoutValue := req.URL.Query().Get("cmdTimeout")

	if timeoutValue == "" {
		klog.Infof("Url Param 'cmdTimeout' is missing, setting default to: %v", cmdExecTimeoutSecond)
	} else {
		timeoutSeconds, err = strconv.Atoi(timeoutValue)
		klog.Infof("Setting timeout to: %v", timeoutSeconds)
		if err != nil {
			klog.Infof("cannot convert timeout param: %v to integer", timeoutValue)
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	httpCommand, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	cmdString := string(httpCommand)
	klog.Infof("running command: %v", cmdString)
	posh := New()
	stdOut, stdErr, err := posh.execute(timeoutSeconds, cmdString)
	klog.Infof("ElevateProcessCmds:\nStdOut : '%s'\nStdErr: '%s'\nErr: %s", strings.TrimSpace(stdOut), stdErr, err)

	if err != nil {
		klog.Infof("Failed to run cmd: %v\n stderr: %v \n stdout: %v", cmdString, stdErr, stdOut)
		resp.WriteHeader(http.StatusBadRequest)
		_, writeErr := resp.Write(append([]byte(stdErr), []byte(stdOut)...))
		if writeErr != nil {
			klog.Info("Failed to write std err and std out to response")
		}
		return
	}

	klog.Infof("cmd: %v success", cmdString)
	resp.WriteHeader(http.StatusOK)
	_, writeErr := resp.Write([]byte(stdOut))
	if writeErr != nil {
		klog.Info("Failed to write std err and std out to response")
	}
}

func httpHealthHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
	_, writeErr := resp.Write([]byte("Stub is Running!"))
	if writeErr != nil {
		klog.Info("Failed to write stub running status to response")
	}
}

func httpFileServeHandler(resp http.ResponseWriter, req *http.Request) {
	httpFileLocation, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fileLocationString := string(httpFileLocation)

	klog.Infof("File to open: %v", fileLocationString)

	if _, err := os.Stat(fileLocationString); err == nil {
		fileBytes, err := ioutil.ReadFile(fileLocationString)
		if err != nil {
			panic(err)
		}
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/octet-stream")
		_, writeErr := resp.Write(fileBytes)
		if writeErr != nil {
			klog.Info("Failed to write fileBytes to response")
		}
		return
	} else if errors.Is(err, os.ErrNotExist) {
		klog.Infof("File does not exist: %v", fileLocationString)
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(resp, "File does not exist: "+fileLocationString+"\n")
	} else {
		klog.Infof("File may not exist: %v", fileLocationString)
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(resp, "File may not exist: "+fileLocationString+"\n")
	}
}

type PowerShell struct {
	powerShell string
}

// New create new session
func New() *PowerShell {
	ps, _ := exec.LookPath("powershell.exe")
	return &PowerShell{
		powerShell: ps,
	}
}

func (p *PowerShell) execute(timeoutSeconds int, args ...string) (stdOut string, stdErr string, err error) {
	args = append([]string{"-NoProfile", "-NonInteractive"}, args...)
	cmd := exec.Command(p.powerShell, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		klog.Infof("Failed to start PowerShell command, args: %v", args)
	}

	// Use a channel to signal completion so we can use a select statement
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// Start a timer
	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)

	// The select statement allows us to execute based on which channel
	// we get a message from first.
	select {
	case <-timeout:
		// Timeout happened first, kill the process and print a message.
		if pKillErr := cmd.Process.Kill(); pKillErr != nil {
			klog.Infof("Failed to kill the process after timeout, error: %v", err)
		}

		klog.Infof("Command timed out")
		err = fmt.Errorf("command timed out")
	case err = <-done:
		// Command completed before timeout. Print output and error if it exists.
		// fmt.Println("Output:", stdout.String())
		if err != nil {
			klog.Infof("Non-zero exit code: %v", err)
		}
	}

	stdOut, stdErr = stdout.String(), stderr.String()
	return
}
