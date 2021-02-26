package server

import "github.com/waittttting/cRPC-common/tcp"

type serviceConn struct {

	serviceName string
	serviceVersion string
	conn *tcp.Connection
}

