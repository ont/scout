package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

const (
	MODE_REQUESTS  = iota
	MODE_RESPONSES = iota
)

// httpStream will handle the actual decoding of http requests.
type HttpStream struct {
	mode           int
	net, transport gopacket.Flow

	reader tcpreader.ReaderStream
	buf    *bufio.Reader

	dumper *Dumper
}

func (hs *HttpStream) Run() {
	//whole, err := ioutil.ReadAll(&hs.reader)
	//if err != nil {
	//	log.Println("Error reading whole stream:", err)
	//	return
	//}
	hs.buf = bufio.NewReader(&hs.reader)
	//buf := bufio.NewReader(bytes.NewReader(whole))

	bytes, err := hs.buf.Peek(4)
	if err != nil {
		log.Println("Error during hs.buf.Peek(4):", err)
		hs.Stop()
		return
	}

	// TODO: better detection of request/response?
	if string(bytes) == "HTTP" {
		hs.mode = MODE_RESPONSES
	} else {
		hs.mode = MODE_REQUESTS
	}

	hs.net.FastHash()
	for {
		switch hs.mode {
		case MODE_REQUESTS:
			req, body, err := hs.ReadRequest()

			if req != nil {
				hs.dumper.Request(hs.net, hs.transport, req, body)
			}

			if err != nil {
				hs.Stop()
				//log.Println("whole:", string(whole))
				//bytes := tcpreader.DiscardBytesToEOF(buf)
				//log.Println("rest:", string(bytes))
				return
			}
			//log.Println("(req) ---->", req)
			//log.Println("(body) --->", string(body))

		case MODE_RESPONSES:
			res, body, err := hs.ReadResponse()

			if res != nil {
				hs.dumper.Response(hs.net, hs.transport, res, body)
			}

			if err != nil {
				hs.Stop()
				//log.Println("whole:", string(whole))
				//bytes := tcpreader.DiscardBytesToEOF(buf)
				//log.Println("rest:", string(bytes))
				return
			}
			//log.Println("(res) <----", res)
			//log.Println("(body) <---", string(body))
		}
	}
}

func (hs *HttpStream) ReadRequest() (*http.Request, []byte, error) {
	req, err := http.ReadRequest(hs.buf)
	if err != nil {
		if !(err == io.EOF || err == io.ErrUnexpectedEOF) {
			log.Println("Error reading request", hs.net, hs.transport, ":", err)
		}
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(req.Body)

	if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {
		body = hs.Decompress(body)
	}

	if err != nil {
		log.Println("Error reading request body", hs.net, hs.transport, ":", err)
		return req, nil, err
	}

	return req, body, nil
}

func (hs *HttpStream) Decompress(body []byte) []byte {
	buf := bytes.NewBuffer(body)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		log.Println("Error creating gzip reader:", err)
		return body
	}

	bytes, err := ioutil.ReadAll(zr)
	if err != nil {
		log.Println("Error during gzip decompressing:", err)
		return body
	}

	return bytes
}

func (hs *HttpStream) ReadResponse() (*http.Response, []byte, error) {
	res, err := http.ReadResponse(hs.buf, nil)
	if err != nil {
		if !(err == io.EOF || err == io.ErrUnexpectedEOF) {
			log.Println("Error reading response", hs.net, hs.transport, ":", err)
		}
		return nil, nil, err
	}

	res.Uncompressed = true
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		if !(err == io.EOF || err == io.ErrUnexpectedEOF) {
			//spew.Dump("Error reading response body", hs.net, hs.transport, err, body, res)
			log.Println("Error reading response body", hs.net, hs.transport, ":", err)
			return res, nil, err
		}
		// we return body as-is but signal about error (for stream termination)
		return res, body, err
	}

	return res, body, nil
}

func (hs *HttpStream) Stop() {
	tcpreader.DiscardBytesToEOF(hs.buf)
}
