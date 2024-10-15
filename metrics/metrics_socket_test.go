package metrics_test

import (
	"fmt"
	"net"
	"strconv"
)

var DfltHost = "0.0.0.0"

type Addr struct {
	Listener net.Listener
	Host     string
	HTTP     string
	Port     int
}

func rndAddr() (*Addr, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("invalid address type: %T", listener.Addr())
	}

	return &Addr{
		Listener: listener,
		Host:     DfltHost,
		Port:     addr.Port,
		HTTP:     fmt.Sprintf("http://%s", net.JoinHostPort(DfltHost, strconv.Itoa(addr.Port))),
	}, nil
}
