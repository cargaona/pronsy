package tcp

import (
	"dns-proxy/pkg/domain/proxy"
	"net"
	"sync"
	"sync/atomic"
)

func NewTCPHandler(packetSize int, logger Logger, cache proxy.Cache, parser proxy.DNSParser) *TCPHandler {
	return &TCPHandler{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, packetSize)
			},
		},
		log:    logger,
		cache:  cache,
		parser: parser,
	}
}

type TCPHandler struct {
	log        Logger
	bufferPool sync.Pool
	cache      proxy.Cache
	parser     proxy.DNSParser
}

func (d *TCPHandler) HandleTCPConnection(conn *net.Conn, p proxy.Service, conns *uint64) {
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
		atomic.AddUint64(conns, ^uint64(0))
		return
	}

	response, err := p.SolveTCP(msg[:nbytes])
	if err != nil {
		d.log.Err("%v", err)
	}
	(*conn).Write(response)

	// Once we write the response to the client save the record in the cache.
	if d.StoreRecordInCache(response) != nil {
		d.log.Err("%v", err)
	}
	atomic.AddUint64(conns, ^uint64(0))
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
