package main

import (
	"flag"
	"github.com/ipconflict/pkg/arp"
	"log"
	"net"
	"net/netip"
	"os"
)

var iface = flag.String("iface", "", "interface")
var netIP = flag.String("ip", "", "target ip address")

const (
	StateNoIPConflict = iota
	StateIPConflict
	StateError
)

func main() {
	flag.Parse()

	if *iface == "" {
		log.Println("interface can't be empty")
		os.Exit(1)
	}

	if *netIP == "" {
		log.Println("ip can't be empty")
		os.Exit(1)
	}

	netIface, err := net.InterfaceByName(*iface)
	if err != nil {
		log.Fatalln(err)
	}

	addr, err := netip.ParseAddr(*netIP)
	if err != nil || !addr.IsValid() {
		log.Fatalln(err)
	}

	state, err := arp.ARPChecking(netIface, addr, addr)
	if err != nil {
		log.Println("failed to checking ipv4 ip address: ", addr.String())
		return
	}

	log.Println("state: ", state)

	switch state {
	case StateIPConflict:
		log.Println("found ip conflict: ", addr.String())
	case StateNoIPConflict:
		log.Printf("ip %s has no conflict", addr.String())
	default:
		log.Println("failed to checking ipv4 ip address")
	}
}
