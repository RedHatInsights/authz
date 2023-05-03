package api

import "net"

type InProcTransport interface {
	net.Listener
	Dial() (net.Conn, error)
}
