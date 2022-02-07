package tcp

import (
	"dns-proxy/pkg/domain/proxy"
	"net"
	"sync"
	"sync/atomic"
)

// NewTCPHandler returns a TCPHandler
func NewTCPHandler(packetSize int, logger Logger, cache proxy.Cache, parser proxy.DNSParser) *TCPHandler {
	return &TCPHandler{
		log:    logger,
		cache:  cache,
		parser: parser,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, packetSize)
			},
		},
	}
}

// TCPHandler has the attributes required for managing the TCP connections, the bufferPool needed to read messages from the requests and things like logger.
type TCPHandler struct {
	log        Logger
	bufferPool sync.Pool
	cache      proxy.Cache
	parser     proxy.DNSParser
}

// HandleTCPConnection reads a message from the server and execute the DNS resolution calling the Proxy service.
func (d *TCPHandler) HandleTCPConnection(conn *net.Conn, p proxy.Service) {
	// TODO: This first block of code that is reading the TCP socket can be moved to a different function.
	defer (*conn).Close()
	msg := d.bufferPool.Get().([]byte)
	nbytes, err := (*conn).Read(msg[:])
	if err != nil {
		d.log.Err("%v", err)
	}

	// Look for message in the cache before resolve it.
	cachedMessage, err := d.GetRecordFromCache(msg)
	if cachedMessage != nil {
		(*conn).Write(cachedMessage)
		atomic.AddUint64(&connections, ^uint64(0))
		return
	}

	// if cachedMessage it's empty go resolve the DNS.
	response, err := p.SolveTCP(msg[:nbytes])
	if err != nil {
		d.log.Err("%v", err)
	}
	(*conn).Write(response)
	// After writing the response the connection is deducted from the connections counter.
	atomic.AddUint64(&connections, ^uint64(0))

	// Once we write the response to the client save the record in the cache.
	if d.StoreRecordInCache(response) != nil {
		d.log.Err("%v", err)
	}
}

func (h *TCPHandler) StoreRecordInCache(msg []byte) error {
	dnsmessage, err := h.parser.TCPMsgToDNS(msg)
	if err != nil {
		return err
	}
	if h.cache.Store(*dnsmessage) != nil {
		return err
	}
	return nil
}

func (h *TCPHandler) GetRecordFromCache(msg []byte) ([]byte, error) {
	dnsmessage, _ := h.parser.TCPMsgToDNS(msg)
	cachedMessage, err := h.cache.Get(*dnsmessage)
	if err != nil {
		h.log.Err("Cache error: %v", err)
	}
	if cachedMessage != nil {
		h.log.Debug("Message found in cache")
		cachedMessage.Header.ID = dnsmessage.Header.ID
		// Message from cache returns as 'dnsmessage'. Parse to 'raw' to return.
		cachedMessageRaw, err := h.parser.DNSToMsg(cachedMessage, proxy.SocketTCP)
		if err != nil {
			return nil, err
		}
		return cachedMessageRaw, nil
	}
	return nil, nil
}
