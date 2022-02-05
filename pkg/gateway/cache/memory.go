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
	ttl    time.Duration
	mx     sync.RWMutex
	logger proxy.Logger
	items  map[string]value
}

func (c *Cache) AutoPurge() {
	for now := range time.Tick(time.Second) {
		for key, value := range c.items {
			if value.expiration.Before(now) {
				c.mx.Lock()
				c.logger.Info("Clearing entry: %v \n", key)
				delete(c.items, key)
				c.mx.Unlock()
			}
		}
	}
}

func New(ttl time.Duration, logger proxy.Logger) proxy.Cache {
	cache := &Cache{
		ttl:    ttl,
		items:  map[string]value{},
		logger: logger,
	}
	return cache
}

func (c *Cache) Get(dnsm dnsmessage.Message) (*dnsmessage.Message, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	c.logger.Debug("Looking for record: %v \n", dnsm.Questions[0].Name)
	if value, ok := c.items[hashKey(dnsm)]; ok {
		c.logger.Debug("Found record: %v \n", dnsm.Questions[0].Name)
		return value.msg, nil
	}
	c.logger.Debug("Record: %v not found", dnsm.Questions[0].Name)
	return nil, nil
}

func (c *Cache) Store(dnsm dnsmessage.Message) error {
	c.mx.Lock()
	c.logger.Debug("Saving record: %v \n", dnsm.Questions[0].Name)
	c.items[hashKey(dnsm)] = value{&dnsm, time.Now().Add(c.ttl)}
	c.mx.Unlock()
	return nil
}

func hashKey(dnsm dnsmessage.Message) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", dnsm.Questions)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
