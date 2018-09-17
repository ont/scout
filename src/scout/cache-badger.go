package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dgraph-io/badger"
	"github.com/google/gopacket"
)

type CacheBadger struct {
	db *badger.DB
}

func NewCacheBadger(path string) *CacheBadger {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.Truncate = true // solves error "Value log truncate required to run DB. This might result in data loss."
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal("Error during creation badger-cache:", err)
	}

	return &CacheBadger{
		db: db,
	}
}

func (c *CacheBadger) CleanEvery(dur time.Duration) {
	for range time.Tick(dur) {
		c.db.RunValueLogGC(0.7) // garbadge collect files which contains more than 70% of garbage
	}
}

func (c *CacheBadger) Request(net, transport gopacket.Flow, req *http.Request, body []byte) error {
	var pair ReqResPair
	pair.SetFlow(net, transport).SetRequest(req, body)
	return c.save(c.key(net, transport), &pair)

}

func (c *CacheBadger) Response(net, transport gopacket.Flow, res *http.Response, body []byte) (*ReqResPair, error) {
	pair, err := c.getOrCreate(net.Reverse(), transport.Reverse())
	if err != nil {
		return nil, err
	}

	pair.SetResponse(res, body)

	// pair with request was loaded from db, now we can delete it from our cache
	if pair.HasRequest {
		err = c.db.Update(func(txn *badger.Txn) error {
			return txn.Delete(c.key(net.Reverse(), transport.Reverse()))
		})
	}

	return pair, err
}

func (c *CacheBadger) getOrCreate(net, transport gopacket.Flow) (*ReqResPair, error) {
	var pair ReqResPair

	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(c.key(net, transport))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if err == badger.ErrKeyNotFound {
			// return new empty pair with saved flow
			pair = ReqResPair{}
			pair.SetFlow(net, transport)

		} else {
			// deserialize data from badger's item value
			val, err := item.Value()
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer(val)
			dec := gob.NewDecoder(buf)
			err = dec.Decode(&pair)
			if err != nil {
				spew.Dump(pair, val, err)
				return err
			}
		}
		return nil
	})

	return &pair, err
}

func (c *CacheBadger) save(key []byte, pair *ReqResPair) error {
	return c.db.Update(func(txn *badger.Txn) error {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(pair); err != nil {
			return err
		}

		// NOTE: TTL of record is 30 minutes for autocleaning dead connections
		// TODO: make it configurable
		return txn.SetWithTTL(key, buf.Bytes(), 30*time.Minute)
	})
}

func (c *CacheBadger) key(net, transport gopacket.Flow) []byte {
	client := append(net.Src().Raw(), ':')
	client = append(client, transport.Src().Raw()...)

	server := append(net.Dst().Raw(), ':')
	server = append(client, transport.Src().Raw()...)

	res := append(client, '|')
	return append(res, server...)
}
