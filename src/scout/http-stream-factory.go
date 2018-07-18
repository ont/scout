package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type HttpStreamFactory struct {
	dumper *Dumper
}

// New is called by assembler for each new stream
func (h *HttpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hs := &HttpStream{
		net:       net,
		transport: transport,
		reader:    tcpreader.NewReaderStream(),

		dumper: h.dumper,
	}
	go hs.Run() // Important... we must guarantee that data from the reader stream is read.

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hs.reader
}
