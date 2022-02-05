package udp

import (
	"dns-proxy/pkg/domain/proxy"
	"net"
	"sync"
	"sync/atomic"
)

// UDPHandler has the attributes required for managing the UDP message queue, the bufferPool needed to read messages from the requests and things like logger.
type UDPHandler struct {
	cache        proxy.Cache
	parser       proxy.DNSParser
	log          Logger
	maxQueueSize int
	messageQueue chan message
	bufferPool   sync.Pool
	ops          uint64
}

// message is the message that travels within the queue.
type message struct {
	addr   net.Addr
	msg    []byte
	length int
	conn   *net.PacketConn
}

// NewUDPHandler returns a UDPHandler
func NewUDPHandler(packetSize, maxQueueSize int, logger Logger, cache proxy.Cache, parser proxy.DNSParser) *UDPHandler {
	return &UDPHandler{
		parser:       parser,
		cache:        cache,
		log:          logger,
		maxQueueSize: maxQueueSize,
		messageQueue: make(chan message, maxQueueSize),
		bufferPool: sync.Pool{
			New: func() interface{} { return make([]byte, packetSize) },
		},
		ops: 0,
	}
}

// handleMessage receives a message from the queue and execute the DNS resolution calling the Proxy service.
func (u *UDPHandler) handleMessage(c net.PacketConn, m *message, p proxy.Service) {
	// Look for message in the cache before resolve it.
	cachedMessage, err := u.GetRecordFromCache(m.msg)
	if cachedMessage != nil {
		_, err = c.WriteTo(cachedMessage[:], m.addr)
		if err != nil {
			u.log.Err("%v", err)
		}
		atomic.AddUint64(&u.ops, 1)
		return
	}
	// if cachedMessage it's empty go resolve the DNS.
	response, err := p.SolveUDP(m.msg)
	if err != nil {
		u.log.Err("%v", err)
	}
	_, err = c.WriteTo(response[:], m.addr)
	if err != nil {
		u.log.Err("%v", err)
	}
	// Once we write the response to the client save the record in the cache.
	if u.StoreRecordInCache(response) != nil {
		u.log.Err("%v", err)
	}

	atomic.AddUint64(&u.ops, 1)
}

func (u *UDPHandler) StoreRecordInCache(msg []byte) error {
	dnsmessage, err := u.parser.UDPMsgToDNS(msg)
	if err != nil {
		return err
	}
	if u.cache.Store(*dnsmessage) != nil {
		return err
	}
	return nil
}

func (u *UDPHandler) GetRecordFromCache(msg []byte) ([]byte, error) {
	dnsmessage, _ := u.parser.UDPMsgToDNS(msg)
	cachedMessage, err := u.cache.Get(*dnsmessage)
	if err != nil {
		u.log.Err("Cache error: %v", err)
	}
	if cachedMessage != nil {
		u.log.Debug("Message found in cache")
		cachedMessage.Header.ID = dnsmessage.Header.ID
		// Message from cache returns as 'dnsmessage'. Parse to 'raw' to return.
		cachedMessageRaw, err := u.parser.DNSToMsg(cachedMessage, proxy.SocketUDP)
		if err != nil {
			return nil, err
		}
		return cachedMessageRaw, nil
	}
	return nil, nil
}

// Dequeue gets the message from the queue and send them to the handler to get the job done.
func (u *UDPHandler) Dequeue(p proxy.Service) {
	for m := range u.messageQueue {
		u.handleMessage(*m.conn, &m, p)
		u.bufferPool.Put(m.msg)
	}
}

// Receive gets the packets comming from the listener, reads them, and sends them to the queue.
func (u *UDPHandler) Receive(c net.PacketConn) {
	defer c.Close()

	for {
		msg := u.bufferPool.Get().([]byte)
		nbytes, addr, err := c.ReadFrom(msg[0:])
		if err != nil {
			u.log.Err("%v", err)
			continue
		}
		u.enqueue(
			message{
				addr,
				msg,
				nbytes,
				&c,
			})
	}
}

// enqueue puts a message in the queue.
func (u *UDPHandler) enqueue(m message) {
	u.messageQueue <- m
}

func (u *UDPHandler) GetOps() uint64 {
	return u.ops
}

func (u *UDPHandler) GetQueueMax() int {
	return u.maxQueueSize
}
