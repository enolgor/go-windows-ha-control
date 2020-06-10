package server

import (
	"fmt"
	"net"
	"os/exec"
)

func Start(port string) error {
	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", ":"+port)
	if err != nil {
		return err
	}
	defer pc.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(pc, addr, buf[:n])
	}
}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	err := doExec(string(buf))
	if err != nil {
		pc.WriteTo([]byte(err.Error()), addr)
	} else {
		pc.WriteTo([]byte("OK"), addr)
	}
}

func doExec(operation string) error {
	var args string
	switch operation {
	case "hibernate":
		args = "shutdown /h /f"
	case "check":
		return nil
	default:
		return fmt.Errorf("Unknown operation")
	}
	return exec.Command("cmd", "/C", args).Run()
}
