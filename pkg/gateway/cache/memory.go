package cache

import (
	"crypto/sha256"
	"dns-proxy/pkg/domain/proxy"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type value struct {
	msg        *dnsmessage.Message
	expiration time.Time
}

type Cache struct {
	enabled bool
	ttl     time.Duration
	mx      sync.Mutex
	log     proxy.Logger
	items   map[string]value
}

func New(ttl time.Duration, logger proxy.Logger, enabled bool) proxy.Cache {
	cache := &Cache{
		ttl:     ttl,
		items:   map[string]value{},
		log:     logger,
		enabled: enabled,
	}
	return cache
}

func (c *Cache) Flush() {
	for now := range time.Tick(time.Second) {
		for key, value := range c.items {
			if value.expiration.Before(now) {
				c.mx.Lock()
				c.log.Info("Clearing entry: %v \n", key)
				delete(c.items, key)
				c.mx.Unlock()
			}
		}
	}
}

func (c *Cache) Get(msg dnsmessage.Message) (*dnsmessage.Message, error) {
	if !c.enabled {
		return nil, nil
	}
	c.mx.Lock()
	defer c.mx.Unlock()
	c.log.Debug("Looking for record: %v \n", msg.Questions[0].Name)
	if value, ok := c.items[hasher(msg)]; ok {
		c.log.Debug("Found record: %v \n", msg.Questions[0].Name)
		return value.msg, nil
	}
	c.log.Debug("Record: %v not found", msg.Questions[0].Name)
	return nil, nil
}

func (c *Cache) Store(msg dnsmessage.Message) error {
	if !c.enabled {
		return nil
	}
	c.mx.Lock()
	c.log.Debug("Saving record: %v \n", msg.Questions[0].Name)
	c.items[hasher(msg)] = value{&msg, time.Now().Add(c.ttl)}
	c.mx.Unlock()
	return nil
}

func hasher(msg dnsmessage.Message) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", msg.Questions)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
