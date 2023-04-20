package arp

import (
	"context"
	"github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
	"log"
	"net"
	"net/netip"
)

const (
	StateNoIPConflict = iota
	StateIPConflict
	StateError
)

func ARPChecking(ifi *net.Interface, sourceIP, targetIP netip.Addr) (int, error) {
	client, err := arp.Dial(ifi)
	if err != nil {
		return StateError, err
	}

	state := StateNoIPConflict
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start a goroutine to receive arp response
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				packet, _, err := client.Read()
				if err != nil {
					continue
				}

				log.Println("receive packet senderip: ", packet.SenderIP.String())
				log.Println("receive packet targetIP: ", packet.TargetIP.String())

				if packet.Operation == arp.OperationReply {
					// found reply and simple check if the reply packet is we want.
					if packet.SenderIP.Compare(targetIP) == 0 {
						state = StateIPConflict
						cancel()
						return
					}
				}
			}
		}
	}()

	// we send a gratuitous arp to checking if ip is conflict
	packet, err := arp.NewPacket(arp.OperationRequest, ifi.HardwareAddr, sourceIP, ethernet.Broadcast, targetIP)
	if err != nil {
		cancel()
		return StateError, err
	}

	// try to send 3 times
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return state, nil
		default:
			err = client.WriteTo(packet, ethernet.Broadcast)
			if err != nil {
				cancel()
				return StateError, err
			}
		}
	}

	return state, nil
}
