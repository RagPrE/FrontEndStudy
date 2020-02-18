package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/network"
)

func PrintPeers(stream network.Stream) {
	fmt.Println("-----RemoteAddr--------")
	fmt.Println(stream.Conn().RemoteMultiaddr().String())
	fmt.Println(stream.Conn().RemotePeer().String())
	fmt.Println("-----My--------")
	fmt.Println(stream.Conn().LocalPeer().String())
	fmt.Println(stream.Conn().LocalMultiaddr().String())
	fmt.Println(idht.RoutingTable().ListPeers())
}
