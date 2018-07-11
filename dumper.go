package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/gopacket"
)

type Dumper struct {
	file  *os.File
	conns map[string]*ReqResPair

	lock sync.RWMutex
}

type ReqResPair struct {
	req *http.Request
	res *http.Response

	reqBody, resBody []byte

	net, tran gopacket.Flow
}

func NewDumper(file *os.File) *Dumper {
	return &Dumper{
		file:  file,
		conns: make(map[string]*ReqResPair),

		lock: sync.RWMutex{},
	}
}

// Request registers new http request (this method probably called from stream)
func (d *Dumper) Request(net, transport gopacket.Flow, req *http.Request, body []byte) {
	d.lock.Lock()
	defer d.lock.Unlock()

	pair := d.FindOrCreate(net, transport)

	pair.req = req
	pair.reqBody = body
}

// Response registers new http response (this method probably called from stream)
func (d *Dumper) Response(net, transport gopacket.Flow, res *http.Response, body []byte) {
	d.lock.Lock()
	defer d.lock.Unlock()

	pair := d.FindOrCreate(net.Reverse(), transport.Reverse())

	pair.res = res
	pair.resBody = body

	d.dump(pair)

	delete(d.conns, d.Key(net.Reverse(), transport.Reverse()))
}

func (d *Dumper) FindOrCreate(net, tran gopacket.Flow) *ReqResPair {
	key := d.Key(net, tran)

	if pair, found := d.conns[key]; found {
		return pair
	} else {
		pair := &ReqResPair{net: net, tran: tran}
		d.conns[key] = pair
		return pair
	}
}

func (d *Dumper) Key(net, tran gopacket.Flow) string {
	client := net.Src().String() + ":" + tran.Src().String()
	server := net.Dst().String() + ":" + tran.Dst().String()
	return client + "|" + server
}

func (d *Dumper) dump(pair *ReqResPair) {
	data := map[string]interface{}{
		"src":      pair.net.Src().String(),
		"dst":      pair.net.Dst().String(),
		"src_port": pair.tran.Src().String(),
		"dst_port": pair.tran.Dst().String(),
	}

	if pair.req != nil {
		data["req"] = map[string]interface{}{
			"method":  pair.req.Method,
			"url":     pair.req.RequestURI,
			"host":    pair.req.Host,
			"headers": pair.req.Header,
			"body":    string(pair.reqBody),
		}
	}

	if pair.res != nil {
		data["res"] = map[string]interface{}{
			"code":    pair.res.StatusCode,
			"status":  pair.res.Status,
			"headers": pair.res.Header,
			"body":    string(pair.resBody),
		}
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Can't json.Marshal data:", err)
		return
	}

	if d.file != nil {
		d.file.Write(bytes)
		d.file.WriteString("\n")
	} else {
		fmt.Println(string(bytes)) // else output to stdout
	}
}
