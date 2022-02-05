package main

import (
	"dns-proxy/internal/config"
	"dns-proxy/pkg/domain/proxy"
	"dns-proxy/pkg/gateway/cache"
	"dns-proxy/pkg/gateway/logger"
	"dns-proxy/pkg/gateway/parser"
	"dns-proxy/pkg/gateway/resolver"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	flushInterval = time.Duration(1) * time.Second
	maxQueueSize  = 1000000
	UDPPacketSize = 1500
)

var address string
var bufferPool sync.Pool
var ops uint64 = 0
var total uint64 = 0
var flushTicker *time.Ticker
var nbWorkers int

func init() {
	flag.StringVar(&address, "addr", ":5353", "Address of the UDP server to test")
	flag.IntVar(&nbWorkers, "concurrency", runtime.NumCPU(), "Number of workers to run in parallel")
}

type message struct {
	addr   net.Addr
	msg    []byte
	length int
	conn   *net.PacketConn
}

type messageQueue chan message

func (mq messageQueue) enqueue(m message) {
	mq <- m
}

func (mq messageQueue) dequeue(p proxy.Service) {
	for m := range mq {
		handleMessage(m.addr, m.msg[0:m.length], p, *m.conn)
		bufferPool.Put(m.msg)
	}
}

var mq messageQueue

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	cfg := config.GetConfig()
	// Create and start cache autopurge.
	cache := cache.New(
		time.Duration(cfg.CacheTTL)*time.Second,
		logger.New("CACHE", true),
	)
	// Create DNS Proxy injecting dependencies.
	proxySvc := proxy.NewDNSProxy(
		resolver.NewResolver(cfg.DnsHost, 853, cfg.ResolverReadTo),
		parser.NewDNSParser(),
		cache,
		logger.New("PROXY", true),
	)

	bufferPool = sync.Pool{
		New: func() interface{} { return make([]byte, UDPPacketSize) },
	}
	mq = make(messageQueue, maxQueueSize)
	listenAndReceive(nbWorkers, proxySvc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			atomic.AddUint64(&total, ops)
			log.Printf("Total ops %d", total)
			os.Exit(0)
		}
	}()

	flushTicker = time.NewTicker(flushInterval)
	for range flushTicker.C {
		log.Printf("Ops/s %f", float64(ops)/flushInterval.Seconds())
		atomic.AddUint64(&total, ops)
		atomic.StoreUint64(&ops, 0)
	}
}

func listenAndReceive(maxWorkers int, p proxy.Service) error {
	c, err := net.ListenPacket("udp", address)
	if err != nil {
		return err
	}
	for i := 0; i < maxWorkers; i++ {
		fmt.Println(address)
		go mq.dequeue(p)
		go receive(c)
	}
	return nil
}

// receive accepts incoming datagrams on c and calls handleMessage() for each message
func receive(c net.PacketConn) {
	defer c.Close()

	for {
		msg := bufferPool.Get().([]byte)
		nbytes, addr, err := c.ReadFrom(msg[0:])
		if err != nil {
			log.Printf("Error %s", err)
			continue
		}
		mq.enqueue(message{addr, msg, nbytes, &c})
	}
}

func handleMessage(addr net.Addr, msg []byte, p proxy.Service, c net.PacketConn) {
	solvedMsg, err := p.SolveUDP(msg)
	if err != nil {
		fmt.Printf("Error solving message: %v \n", err)
	}

	_, err = c.WriteTo(solvedMsg[:], addr)
	if err != nil {
		fmt.Println(err)
	}
	// Do something with message
	atomic.AddUint64(&ops, 1)
}
