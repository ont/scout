package main

import (
	"net/http"

	"github.com/google/gopacket"
)

type Cache interface {
	Request(net, transport gopacket.Flow, req *http.Request, body []byte) error
	Response(net, transport gopacket.Flow, res *http.Response, body []byte) (*ReqResPair, error)
}
