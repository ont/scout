package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

var (
	iface   = kingpin.Flag("eth0", "Interface to get packets from").Short('i').String()
	fin     = kingpin.Flag("in", "PCAP file to read from, overrides --eth0").String()
	fout    = kingpin.Flag("out", "Filename to write to").OpenFile(os.O_CREATE|os.O_RDWR, 0660)
	snaplen = kingpin.Flag("s", "How many bytes will be caputured for each packet").Default("262144").Int()
	timeout = kingpin.Flag("timeout", "Flush inactive connections after this amount of minutes (for live capturing).").Default("20").Int()
	filter  = kingpin.Flag("filter", "BPF filter for pcap").Short('F').Default("tcp and port 80").String()
	verbose = kingpin.Flag("verbose", "Print each raw captured packet").Short('v').Bool()
)

/*
 * NOTE: code is heavly based on standart gopacket httpassembly example
 */
func main() {
	kingpin.Parse()

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

	// Set up assembly
	streamFactory := &HttpStreamFactory{
		dumper: NewDumper(*fout),
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
