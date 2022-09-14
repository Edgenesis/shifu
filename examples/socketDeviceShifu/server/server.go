package main

import (
	"net"

	"k8s.io/klog/v2"
)

func main() {
	addr := "0.0.0.0:11122"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	klog.Infoln("listening at ", addr)
	for {
		conn, err := listener.Accept()
		klog.Infoln(conn.RemoteAddr())
		if err != nil {
			break
		}

		go handleReq(conn)
	}
}
func handleReq(conn net.Conn) {
	for {
		data := make([]byte, 1024)
		_, err := conn.Read(data)
		klog.Infoln(string(data), err)
		if err != nil {
			klog.Errorln(err)
			return
		}

		conn.Write(data)
	}
}
