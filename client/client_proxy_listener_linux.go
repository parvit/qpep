//go:build linux

package client

import (
	"fmt"
	"github.com/parvit/qpep/shared"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// ClientProxyListener implements the local listener for diverted connections
type ClientProxyListener struct {
	base net.Listener
}

// Accept method accepts the connections from generic connection types
func (listener *ClientProxyListener) Accept() (net.Conn, error) {
	return listener.AcceptTProxy()
}

// AcceptTProxy method accepts the connections and casts those to a tcp connection type
func (listener *ClientProxyListener) AcceptTProxy() (*net.TCPConn, error) {
	if listener.base == nil {
		return nil, shared.ErrFailed
	}
	tcpConn, err := listener.base.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return nil, err
	}
	return tcpConn, nil
}

// Addr method returns the listening address
func (listener *ClientProxyListener) Addr() net.Addr {
	if listener.base == nil {
		return nil
	}
	return listener.base.Addr()
}

// Addr method close the listener
func (listener *ClientProxyListener) Close() error {
	if listener.base == nil {
		return nil
	}
	return listener.base.Close()
}

// NewClientProxyListener method instantiates a new ClientProxyListener on a tcp address base listener
func NewClientProxyListener(network string, laddr *net.TCPAddr) (net.Listener, error) {
	//Open basic TCP listener
	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}

	//Find associated file descriptor for listener to set socket options on
	fileDescriptorSource, err := listener.File()
	if err != nil {
		return nil, &net.OpError{Op: "ClientListener", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("get file descriptor: %s", err)}
	}
	defer fileDescriptorSource.Close()

	//Make the port transparent so the gateway can see the real origin IP address (invisible proxy within satellite environment)
	_ = syscall.SetsockoptInt(int(fileDescriptorSource.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	_ = syscall.SetsockoptInt(int(fileDescriptorSource.Fd()), syscall.SOL_TCP, unix.TCP_FASTOPEN, 1)

	//return a derived TCP listener object with TCProxy support
	return &ClientProxyListener{base: listener}, nil
}
