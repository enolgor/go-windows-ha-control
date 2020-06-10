package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/enolgor/go-windows-ha-control/client"
	"github.com/enolgor/go-windows-ha-control/server"
)

var modeStr string

const (
	modeFlag  = "mode"
	modeUsage = "Mode: 'client' to start ha-rest-switch, 'server' to start Hibernate server"
)

var ipAddrStr string

const (
	ipAddrFlag  = "ip"
	ipAddrUsage = "IP Addr to send WOL/Hibernate"
)

var netAddrStr string

const (
	netAddrFlag  = "net"
	netAddrUsage = "Net Broadcast Addr to send WOL"
)

var macAddrStr string

const (
	macAddrFlag  = "mac"
	macAddrUsage = "Mac Addr to send WOL"
)

var clientPortStr string

const (
	clientPortFlag  = "clientport"
	clientPortUsage = "Client TCP Port to listen"
)

var serverPortStr string

const (
	serverPortFlag  = "serverport"
	serverPortUsage = "Server UDP port to send hibernate"
)

var timeout int

const (
	timeoutFlag  = "timeout"
	timeoutUsage = "Timeout for waiting response (seconds)"
)

func init() {
	log.SetOutput(os.Stdout)

	envModeStr := os.Getenv("MODE")
	envIPAddrStr := os.Getenv("IP_ADDR")
	envNetAddrStr := os.Getenv("NET_ADDR")
	envMacAddrStr := os.Getenv("MAC_ADDR")
	envClientPortStr := os.Getenv("CLIENT_PORT")
	envServerPortStr := os.Getenv("SERVER_PORT")
	envTimeoutStr := os.Getenv("TIMEOUT")
	envTimeout := 0
	if envTimeoutStr != "" {
		var err error
		if envTimeout, err = strconv.Atoi(envTimeoutStr); err != nil {
			panic("Env var TIMEOUT is not an int")
		}
	}

	flag.StringVar(&modeStr, modeFlag, envModeStr, modeUsage)
	flag.StringVar(&ipAddrStr, ipAddrFlag, envIPAddrStr, ipAddrUsage)
	flag.StringVar(&netAddrStr, netAddrFlag, envNetAddrStr, netAddrUsage)
	flag.StringVar(&macAddrStr, macAddrFlag, envMacAddrStr, macAddrUsage)
	flag.StringVar(&clientPortStr, clientPortFlag, envClientPortStr, clientPortUsage)
	flag.StringVar(&serverPortStr, serverPortFlag, envServerPortStr, serverPortUsage)
	flag.IntVar(&timeout, timeoutFlag, envTimeout, timeoutUsage)
	flag.Parse()

	switch modeStr {
	case "client":
		if ipAddrStr == "" || netAddrStr == "" || macAddrStr == "" || clientPortStr == "" || serverPortStr == "" || timeout <= 0 {
			flag.PrintDefaults()
			panic("You must specify ip, net, mac, port and timeout > 0")
		}
		log.Printf("Client Mode\nTarget IP=%s; BCast NW=%s; Target MAC=%s; Client TCP Port=%s; Server UDP Port=%s; Timeout=%d\n", ipAddrStr, netAddrStr, macAddrStr, clientPortStr, serverPortStr, timeout)
	case "server":
		if serverPortStr == "" {
			flag.PrintDefaults()
			panic("You must specify UDP port to listen")
		}
		log.Printf("Server Mode\n Listen UDP Port=%s\n", serverPortStr)
	default:
		flag.PrintDefaults()
		panic("Mode should be 'client' or 'server'")
	}
}

func main() {
	var err error
	switch modeStr {
	case "client":
		err = doClient()
	case "server":
		err = doServer()
	}
	if err != nil {
		log.Fatal("FATAL ERROR: ", err)
	}
}

func doClient() error {
	var err error
	var ipAddr *net.IPAddr
	var netAddr *net.UDPAddr
	var mac net.HardwareAddr
	var c *client.WHControlClient
	if ipAddr, err = net.ResolveIPAddr("", "192.168.1.107"); err != nil {
		return err
	}
	if netAddr, err = net.ResolveUDPAddr("", "192.168.1.255:7"); err != nil {
		return err
	}
	if mac, err = net.ParseMAC("D0:17:C2:AB:A6:C9"); err != nil {
		return err
	}
	if c, err = client.NewClient(ipAddr, netAddr, mac, serverPortStr, timeout); err != nil {
		return err
	}
	/*if err = c.Wake(); err != nil {
		return err
	}*/
	/*if err = c.Check(); err != nil {
		return err
	}*/
	haSwitch := client.NewHASwitch(clientPortStr, c)
	if err = haSwitch.Start(); err != nil {
		return err
	}
	return nil
}

func doServer() error {
	return server.Start(serverPortStr)
}
