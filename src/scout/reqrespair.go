package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/gopacket"
)

type ReqResPair struct {
	req *http.Request
	res *http.Response

	reqBody, resBody []byte

	net, tran gopacket.Flow
}

func (p *ReqResPair) AsJson() ([]byte, error) {
	data := map[string]interface{}{
		"src":      p.net.Src().String(),
		"dst":      p.net.Dst().String(),
		"src_port": p.tran.Src().String(),
		"dst_port": p.tran.Dst().String(),
	}

	if p.req != nil {
		data["req"] = map[string]interface{}{
			"method":  p.req.Method,
			"url":     p.req.RequestURI,
			"host":    p.req.Host,
			"headers": p.req.Header,
			"body":    string(p.reqBody),
		}
	}

	if p.res != nil {
		data["res"] = map[string]interface{}{
			"code":    p.res.StatusCode,
			"status":  p.res.Status,
			"headers": p.res.Header,
			"body":    string(p.resBody),
		}
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Can't json.Marshal data:", err)
		return nil, err
	}

	return bytes, nil
}
