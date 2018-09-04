package main

import (
	"net/http"
	"sync"

	"github.com/google/gopacket"
)

type CacheMap struct {
	lock sync.RWMutex

	c map[string]*ReqResPair
}

func NewCacheMap() *CacheMap {
	return &CacheMap{
		c: make(map[string]*ReqResPair),
	}
}

func (c *CacheMap) Request(net, transport gopacket.Flow, req *http.Request, body []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	pair := c.findOrCreate(net, transport)

	pair.SetRequest(req, body)

	return nil
}

func (c *CacheMap) Response(net, transport gopacket.Flow, res *http.Response, body []byte) (*ReqResPair, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	pair := c.findOrCreate(net.Reverse(), transport.Reverse())

	pair.SetResponse(res, body)

	delete(c.c, c.key(net.Reverse(), transport.Reverse()))

	return pair, nil
}

func (c *CacheMap) findOrCreate(net, tran gopacket.Flow) *ReqResPair {
	key := c.key(net, tran)

	if pair, found := c.c[key]; found {
		return pair
	} else {
		pair := &ReqResPair{}
		pair.SetFlow(net, tran)
		c.c[key] = pair
		return pair
	}
}

func (c *CacheMap) key(net, tran gopacket.Flow) string {
	client := net.Src().String() + ":" + tran.Src().String()
	server := net.Dst().String() + ":" + tran.Dst().String()
	return client + "|" + server
}
