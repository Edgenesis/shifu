package rtspRecord

import (
	"fmt"
	"github.com/edgenesis/shifu/pkg/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

// CmdMapMemory map[string]*exec.Cmd, from deviceName to running command, no persistence
var CmdMapMemory sync.Map

type Device struct {
	DeviceName string
	In         string
	Running    bool
	Clip       int
}

func (d *Device) createCmd() *exec.Cmd {
	cmd := d.compile()
	go func() {
		err := cmd.Run()
		if err != nil {
			logger.Error(err)
			return
		}
	}()
	return cmd
}

func (d *Device) startRecord() {
	CmdMapMemory.Store(d.DeviceName, d.createCmd())
	d.Clip += 1
	d.Running = true
}

func (d *Device) stopRecord() error {
	inter, exist := CmdMapMemory.LoadAndDelete(d.DeviceName)
	d.Running = false
	if !exist {
		return fmt.Errorf("process not found for device %v", d.DeviceName)
	}
	if inter.(*exec.Cmd) == nil {
		return fmt.Errorf("nil for device %v", d.DeviceName)
	}
	err := inter.(*exec.Cmd).Process.Signal(syscall.SIGINT)
	if err != nil {
		return err
	}
	return nil
}

func (d *Device) compile() *exec.Cmd {
	out := filepath.Join(VideoSavePath, d.DeviceName+"_"+strconv.Itoa(d.Clip)+".mp4")
	return ffmpeg.Input(d.In, ffmpeg.KwArgs{"rtsp_transport": "tcp"}).
		Output(out, ffmpeg.KwArgs{"c": "copy"}).
		OverWriteOutput().ErrorToStdOut().Compile()
}
