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

var bootTime int

const (
	bootTimeFlag  = "bootTime"
	bootTimeUsage = "Time for waiting reboot"
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
	envBootTimeStr := os.Getenv("BOOT_TIME")
	envBootTime := 0
	if envTimeoutStr != "" {
		var err error
		if envTimeout, err = strconv.Atoi(envTimeoutStr); err != nil {
			panic("Env var TIMEOUT is not an int")
		}
	}
	if envBootTimeStr != "" {
		var err error
		if envBootTime, err = strconv.Atoi(envBootTimeStr); err != nil {
			panic("Env var BOOT_TIME is not an int")
		}
	}

	flag.StringVar(&modeStr, modeFlag, envModeStr, modeUsage)
	flag.StringVar(&ipAddrStr, ipAddrFlag, envIPAddrStr, ipAddrUsage)
	flag.StringVar(&netAddrStr, netAddrFlag, envNetAddrStr, netAddrUsage)
	flag.StringVar(&macAddrStr, macAddrFlag, envMacAddrStr, macAddrUsage)
	flag.StringVar(&clientPortStr, clientPortFlag, envClientPortStr, clientPortUsage)
	flag.StringVar(&serverPortStr, serverPortFlag, envServerPortStr, serverPortUsage)
	flag.IntVar(&timeout, timeoutFlag, envTimeout, timeoutUsage)
	flag.IntVar(&bootTime, bootTimeFlag, envBootTime, bootTimeUsage)
	flag.Parse()

	switch modeStr {
	case "client":
		if ipAddrStr == "" || netAddrStr == "" || macAddrStr == "" || clientPortStr == "" || serverPortStr == "" || timeout <= 0 || bootTime <= 0 {
			flag.PrintDefaults()
			panic("You must specify ip, net, mac, port, timeout > 0 and bootTime > 0")
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
	if ipAddr, err = net.ResolveIPAddr("", ipAddrStr); err != nil {
		return err
	}
	if netAddr, err = net.ResolveUDPAddr("", netAddrStr); err != nil {
		return err
	}
	if mac, err = net.ParseMAC(macAddrStr); err != nil {
		return err
	}
	if c, err = client.NewClient(ipAddr, netAddr, mac, serverPortStr, timeout); err != nil {
		return err
	}
	haSwitch := client.NewHASwitch(clientPortStr, c, bootTime)
	if err = haSwitch.Start(); err != nil {
		return err
	}
	return nil
}

func doServer() error {
	return server.Start(serverPortStr)
}
