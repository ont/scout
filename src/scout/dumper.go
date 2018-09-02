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

	conns map[string]*ReqResPair

	lock sync.RWMutex
}

func NewDumper(out Storage) *Dumper {
	return &Dumper{
		out:   out,
		conns: make(map[string]*ReqResPair),
		lock:  sync.RWMutex{},
	}
}

// Request registers new http request (this method probably called from stream)
func (d *Dumper) Request(net, transport gopacket.Flow, req *http.Request, body []byte) {
	d.lock.Lock()
	defer d.lock.Unlock()

	pair := d.findOrCreate(net, transport)

	pair.req = req
	pair.reqBody = body
}

// Response registers new http response (this method probably called from stream)
func (d *Dumper) Response(net, transport gopacket.Flow, res *http.Response, body []byte) {
	d.lock.Lock()
	defer d.lock.Unlock()

	pair := d.findOrCreate(net.Reverse(), transport.Reverse())

	pair.res = res
	pair.resBody = body

	d.dump(pair)

	delete(d.conns, d.key(net.Reverse(), transport.Reverse()))
}

func (d *Dumper) findOrCreate(net, tran gopacket.Flow) *ReqResPair {
	key := d.key(net, tran)

	if pair, found := d.conns[key]; found {
		return pair
	} else {
		pair := &ReqResPair{net: net, tran: tran}
		d.conns[key] = pair
		return pair
	}
}

func (d *Dumper) key(net, tran gopacket.Flow) string {
	client := net.Src().String() + ":" + tran.Src().String()
	server := net.Dst().String() + ":" + tran.Dst().String()
	return client + "|" + server
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
