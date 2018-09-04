package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/gopacket"
)

type ReqResPair struct {
	HasRequest bool
	Method     string
	Url        string
	Path       string
	Host       string
	ReqHeaders http.Header

	HasResponse bool
	Code        int
	Status      string
	ResHeaders  http.Header

	//Req *http.Request
	//Res *http.Response

	ReqBody, ResBody []byte

	Src, Dst, SrcPort, DstPort string
	//net, tran gopacket.Flow
}

func (p *ReqResPair) SetRequest(req *http.Request, body []byte) *ReqResPair {
	p.HasRequest = true

	p.Method = req.Method
	p.Url = req.RequestURI
	p.Host = req.Host
	p.Path = req.URL.Path
	p.ReqHeaders = req.Header
	p.ReqBody = body

	return p
}

func (p *ReqResPair) SetResponse(res *http.Response, body []byte) *ReqResPair {
	p.HasResponse = true

	p.Code = res.StatusCode
	p.Status = res.Status
	p.ResHeaders = res.Header
	p.ResBody = body

	return p
}

func (p *ReqResPair) SetFlow(net, tran gopacket.Flow) *ReqResPair {
	p.Src = net.Src().String()
	p.Dst = net.Dst().String()
	p.SrcPort = tran.Src().String()
	p.DstPort = tran.Dst().String()

	return p
}

func (p *ReqResPair) Cookie(name string) (*http.Cookie, error) {
	req := http.Request{Header: p.ReqHeaders}
	return req.Cookie(name)
}

func (p *ReqResPair) AsJson() ([]byte, error) {
	data := map[string]interface{}{
		"src":      p.Src,
		"dst":      p.Dst,
		"src_port": p.SrcPort,
		"dst_port": p.DstPort,
	}

	if p.HasRequest {
		data["req"] = map[string]interface{}{
			"method":  p.Method,
			"url":     p.Url,
			"host":    p.Host,
			"headers": p.ReqHeaders,
			"body":    string(p.ReqBody),
		}
	}

	if p.HasResponse {
		data["res"] = map[string]interface{}{
			"code":    p.Code,
			"status":  p.Status,
			"headers": p.ResHeaders,
			"body":    string(p.ResBody),
		}
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Can't json.Marshal data:", err)
		return nil, err
	}

	return bytes, nil
}
