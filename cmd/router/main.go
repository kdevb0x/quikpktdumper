// Copyright (C) 2019-2020 Kdevb0x Ltd.
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

// +build linux mips

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

var dev = "wlan0"
var snaplen = uint32(1024)
var proxyaddr string
var filecount = 0

func init() {
	log.SetPrefix("<packetproxy>: ")
}

func printDevs() {
	d, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}
	for _, dev := range d {
		fmt.Printf("\nName: %s\n Description: %s\n", dev.Name, dev.Description)
		for _, addr := range dev.Addresses {
			fmt.Printf("- IP address: %v\n Subnet mask: %v\n", addr.IP, addr.Netmask)
		}
	}

}

func capture(fname string) {
	f, err := os.Create(fname + fmt.Sprintf("_%d", filecount))
	if err != nil {
		log.Fatalln(err)
	}
	filecount++
	w := pcapgo.NewWriter(f)
	w.WriteFileHeader(snaplen, layers.LinkTypeEthernet)
	defer f.Close()

	handle, err := pcap.OpenLive(dev, int32(65535), false, -1*time.Second)
	if err != nil {
		log.Fatalln(err)

	}
	defer handle.Close()

	src := gopacket.NewPacketSource(handle, handle.LinkType())

	for pkt := range src.Packets() {
		w.WritePacket(pkt.Metadata().CaptureInfo, pkt.Data())
	}

}

func proxyCapture(ctx context.Context, daemonaddr string) {
	conn, err := net.Dial("tcp", daemonaddr)
	if err != nil {
		log.Fatal(err)
	}
	w := pcapgo.NewWriter(conn)
	w.WriteFileHeader(snaplen, layers.LinkTypeEthernet)
	handle, err := pcap.OpenLive(dev, int32(65535), false, -1*time.Second)
	if err != nil {
		log.Fatalln(err)

	}

	defer handle.Close()
	src := gopacket.NewPacketSource(handle, handle.LinkType())

	select {
	case <-ctx.Done():
		log.Println("received context cancelation, closing connection")
		conn.Close()
		return
	default:
		for pkt := range src.Packets() {
			w.WritePacket(pkt.Metadata().CaptureInfo, pkt.Data())
		}
	}
}

func main() {
	flag.StringVar(&proxyaddr, "l", "", "proxy listener address")
	flag.Parse()
	if len(os.Args[1:]) < 1 {

		printDevs()
	}
	if proxyaddr != "" {
		log.Printf("connecting to listener at %q\n", proxyaddr)
		proxyCapture(context.Background(), proxyaddr)
	}

	// capture(os.Args[1])

}
