package tcp

import (
	"dns-proxy/pkg/domain/proxy"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	flushInterval = time.Duration(1) * time.Second
)

var ops uint64 = 0
var total uint64 = 0
var connections uint64 = 0
var flushTicker *time.Ticker

type TCPServer struct {
	handler           HandlerTCP
	proxySvc          proxy.Service
	log               Logger
	port              int
	host              string
	maxPoolConnection int
}

// Logger interface is used to inject different implementations of loggers.
type Logger interface {
	Info(format string, i ...interface{})
	Err(format string, i ...interface{})
	Debug(format string, i ...interface{})
}

type HandlerTCP interface {
	HandleTCPConnection(conn *net.Conn, p proxy.Service)
}

func New(proxy proxy.Service, tcpHandler HandlerTCP, logger Logger, port int, host string, maxPoolConnection int) *TCPServer {
	return &TCPServer{
		proxySvc:          proxy,
		handler:           tcpHandler,
		log:               logger,
		port:              port,
		host:              host,
		maxPoolConnection: maxPoolConnection,
	}
}

func (d *TCPServer) Serve() {
	d.log.Debug("Starting to serve TCP")
	d.log.Debug("Max number of incoming connections  %v", d.maxPoolConnection)
	// Start listener. No blocking.
	d.listendAndReceive()
}

func (d *TCPServer) listendAndReceive() error {
	d.log.Info("### Listening TCP at :%d", d.port)
	ln, err := net.Listen(proxy.SocketTCP, net.JoinHostPort(d.host, strconv.Itoa(d.port)))
	if err != nil {
		d.log.Err("%v", err)
		panic(err)
	}
	for {
		if connections <= uint64(d.maxPoolConnection-1) {
			conn, err := ln.Accept()
			// Add 1 to the connection atomic counter.
			atomic.AddUint64(&connections, 1)
			d.log.Debug("current TCP connections %v", connections)
			if err != nil {
				d.log.Err("%v", err)
				//panic(err)
			}
			go d.handler.HandleTCPConnection(&conn, d.proxySvc)
		}
	}
}
