package tcp

import (
	"dns-proxy/pkg/domain/proxy"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	flushInterval = time.Duration(1) * time.Second
)

var ops uint64 = 0
var total uint64 = 0
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
	HandleTCPConnection(conn *net.Conn, p proxy.Service, conns *uint64)
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
	d.listendAndReceive()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			atomic.AddUint64(&total, ops)
			d.log.Info("Interrupt signal detected. Finishing application")
			os.Exit(0)
		}
	}()

	// TODO: limit the tcp connections
	// TODO: understand ticker
	flushTicker = time.NewTicker(flushInterval)
	for range flushTicker.C {
		atomic.AddUint64(&total, ops)
		atomic.StoreUint64(&ops, 0)
	}
}

func (d *TCPServer) listendAndReceive() error {
	d.log.Info("### Listening TCP at :%d", d.port)
	ln, err := net.Listen(proxy.SocketTCP, d.host+":"+strconv.Itoa(d.port))
	if err != nil {
		d.log.Err("%v", err)
		panic(err)
	}
	var conns uint64
	for {
		if conns <= uint64(d.maxPoolConnection-1) {
			conn, err := ln.Accept()
			atomic.AddUint64(&conns, 1)
			if err != nil {
				d.log.Err("%v", err)
				//panic(err)
			}
			go d.handler.HandleTCPConnection(&conn, d.proxySvc, &conns)
		}
	}
}
