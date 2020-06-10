package client

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/mdlayher/wol"
)

type WHControlClient struct {
	c          *wol.Client
	ipAddr     *net.IPAddr
	netAddr    *net.UDPAddr
	mac        net.HardwareAddr
	serverPort string
	timeout    int
}

func NewClient(ipAddr *net.IPAddr, netAddr *net.UDPAddr, mac net.HardwareAddr, serverPort string, timeout int) (*WHControlClient, error) {
	c, err := wol.NewClient()
	if err != nil {
		return nil, err
	}
	return &WHControlClient{c, ipAddr, netAddr, mac, serverPort, timeout}, nil
}

func (whcc *WHControlClient) Wake() error {
	return whcc.c.Wake(whcc.netAddr.String(), whcc.mac)
}

func (whcc *WHControlClient) Check() error {
	return whcc.sendUDP("check", true)
}

func (whcc *WHControlClient) Hibernate() error {
	return whcc.sendUDP("hibernate", false)
}

func (whcc *WHControlClient) Start(state *State, bootGrace *State) {
	for {
		if !bootGrace.Get() {
			if err := whcc.Check(); err != nil {
				state.Set(false)
			} else {
				state.Set(true)
			}
		}
		time.Sleep(time.Duration(2*whcc.timeout) * time.Second)
	}
}

func (whcc *WHControlClient) sendUDP(operation string, waitResp bool) error {
	p := make([]byte, 1024)
	conn, err := net.Dial("udp", whcc.ipAddr.String()+":"+whcc.serverPort)
	defer conn.Close()
	if err != nil {
		return err
	}
	fmt.Fprintf(conn, operation)
	if !waitResp {
		return nil
	}
	go func() {
		time.Sleep(time.Duration(whcc.timeout) * time.Second)
		conn.Close()
	}()
	var n int
	n, err = bufio.NewReader(conn).Read(p)
	resp := string(p[:n])
	if err == nil {
		if resp == "OK" {
			return nil
		}
		return fmt.Errorf(resp)
	}
	return err
}
