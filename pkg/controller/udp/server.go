package udp

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

var total uint64 = 0
var ops uint64 = 0
var flushTicker *time.Ticker

// Logger interface is used to inject different implementations of loggers.
type Logger interface {
	Info(format string, i ...interface{})
	Err(format string, i ...interface{})
	Debug(format string, i ...interface{})
}

type HandlerUDP interface {
	Receive(net.PacketConn)
	Dequeue(proxy.Service)
	//   GetOps() uint64
	GetQueueMax() int
	//   WriteOps(uint64)
}

// UDPServer is the struct that contains the configuration for the UDP server.
// Its methods allow to listen incoming packets and handle them with it 'handler' object.
type UDPServer struct {
	proxySvc   proxy.Service
	log        Logger
	port       int
	maxWorkers int
	handler    HandlerUDP
}

func New(proxy proxy.Service, udpHandler HandlerUDP, logger Logger, port int, maxWorkers int) *UDPServer {
	return &UDPServer{
		proxySvc:   proxy,
		log:        logger,
		port:       port,
		maxWorkers: maxWorkers,
		handler:    udpHandler,
	}
}

func (d *UDPServer) Serve() {
	d.log.Debug("Starting to serve UDP")
	d.log.Debug("Number of CPUS %v", d.maxWorkers)
	d.log.Debug("Max number of records to solve in queue %d", d.handler.GetQueueMax())
	// Start listener. No blocking.
	d.listenAndReceive()

	// Wait for an 'Interrupt' signal to show the counter and finish the application.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// Add the last operations to the total once the interrupt signal comes,  display the data and close the application.
			atomic.AddUint64(&total, ops)
			d.log.Info("Total Operations %d", total)
			d.log.Info("Interrupt signal detected. Finishing application")
			os.Exit(0)
		}
	}()

	// The ticker prevents this routine to finish while the 'receive' and 'dequeue' functions are reading and processing packets.
	flushTicker = time.NewTicker(flushInterval)
	for range flushTicker.C {
		// Every flushTicker.C time add the operations to the total and reset the operations counter.
		atomic.AddUint64(&total, ops)
		atomic.StoreUint64(&ops, 0)
	}
}

func (s *UDPServer) listenAndReceive() error {
	s.log.Info("### Listening UDP at :%d", s.port)
	c, err := net.ListenPacket(proxy.SocketUDP, ":"+strconv.Itoa(s.port))
	if err != nil {
		s.log.Err("Error listening packets", err)
		return err
	}
	// Spawn maxWorkers number of goroutines that will handle the incoming UDP packets.
	for i := 0; i < s.maxWorkers; i++ {
		go s.handler.Dequeue(s.proxySvc)
		go s.handler.Receive(c)
	}
	return nil
}
