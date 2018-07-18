package main

import (
	"io"
	"log"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"

	_ "net/http/pprof"
)

var (
	iface   = kingpin.Flag("iface", "Interface to get packets from (enables live capturing)").Short('i').Default("eth0").String()
	filter  = kingpin.Flag("filter", "BPF filter for pcap").Short('F').Default("tcp and port 80").String()
	snaplen = kingpin.Flag("s", "How many bytes will be caputured for each packet").Default("262144").Int()
	timeout = kingpin.Flag("timeout", "Flush inactive connections after this amount of minutes (for live capturing).").Default("20").Int()
	fin     = kingpin.Flag("read", "PCAP file to read from, overrides --iface").Short('r').String()
	fout    = kingpin.Flag("write", "Filename to write to request-response pairs (in JSON format)").Short('w').String()
	verbose = kingpin.Flag("verbose", "Print each raw captured packet").Short('v').Bool()
	bags    = kingpin.Flag("bag", "Filer in format \"regexp|filename\" for sorting incoming requests via URL into files. Can be repeated multiple times.").Short('b').Strings()
	rotate  = kingpin.Flag("rotate", "Rotate each output file each N-hours. This option specify N value.").Short('R').Default("0").Int()
	debug   = kingpin.Flag("debug", "Debug port to run go pprof on.").Short('d').String()
)

/*
 * NOTE: code is heavly based on standart gopacket httpassembly example
 */
func main() {
	kingpin.Parse()

	if *debug != "" {
		RunProfiler(*debug)
	}

	var handle *pcap.Handle
	var err error

	// Set up pcap packet capture
	if *fin != "" {
		log.Printf("Reading from pcap dump %q", *fin)
		handle, err = pcap.OpenOffline(*fin)
	} else {
		log.Printf("Starting capture on interface %q", *iface)
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), true, pcap.BlockForever)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := handle.SetBPFFilter(*filter); err != nil {
		log.Fatal(err)
	}

	dur := time.Duration(*rotate) * time.Hour

	var rw io.Writer
	if *fout != "" {
		rw = NewRotateWriter(*fout, dur)
	}

	dumper := NewDumper(rw)

	for _, bag := range *bags {
		parts := strings.Split(bag, "::")

		if len(parts) < 2 {
			log.Fatal("Wrong format for bag, must be in form 'regexp::filename'")
		}

		dumper.AddBag(NewBagRegexp(
			parts[0],
			NewRotateWriter(parts[1], dur),
		))
	}

	// Set up assembly
	streamFactory := &HttpStreamFactory{
		dumper: dumper,
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	log.Println("reading in packets")
	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}
			if *verbose {
				log.Println("captured packet:", packet)
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				log.Println("Unusable packet")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 20 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -time.Duration(*timeout)))
		}
	}
}
