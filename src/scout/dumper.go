package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/gopacket"
)

type Dumper struct {
	out  Storage
	bags []Bag

	cache Cache

	lock sync.RWMutex
}

func NewDumper(out Storage, cache Cache) *Dumper {
	return &Dumper{
		out:   out,
		cache: cache,
		lock:  sync.RWMutex{},
	}
}

// Request registers new http request (this method probably called from stream)
func (d *Dumper) Request(net, transport gopacket.Flow, req *http.Request, body []byte) {
	if err := d.cache.Request(net, transport, req, body); err != nil {
		log.Fatal("Error during request registration:", err)
	}
}

// Response registers new http response (this method probably called from stream)
func (d *Dumper) Response(net, transport gopacket.Flow, res *http.Response, body []byte) {
	if pair, err := d.cache.Response(net, transport, res, body); err != nil {
		log.Fatal("Error during response registration:", err)
	} else {
		d.dump(pair)
	}
}

func (d *Dumper) dump(pair *ReqResPair) {
	d.dumpToOut(pair)
	d.dumpToBags(pair)

}

func (d *Dumper) dumpToOut(pair *ReqResPair) {
	if d.out != nil {
		err := d.out.Save(pair)

		if err != nil {
			log.Fatal("Error during saving pair: ", err)
		}

	} else {
		bytes, err := pair.AsJson()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bytes)) // else output to stdout
	}
}

func (d *Dumper) dumpToBags(pair *ReqResPair) {
	for _, bag := range d.bags {
		err := bag.Write(pair)
		if err != nil {
			log.Fatal("Error during writing to output bag: ", err)
		}
	}
}

func (d *Dumper) AddBag(b Bag) {
	d.bags = append(d.bags, b)
}
