package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	crypto "github.com/libp2p/go-libp2p-crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
	// "github.com/qjpcpu/log"
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

func GetPrivateKey(nodeName string) crypto.PrivKey {
	os.MkdirAll(nodeName, 0777)
	filename := nodeName + "/private-key"
	var pk crypto.PrivKey
	for loop := true; loop; loop = false {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			break
		}
		//pk = key.PrivateKeyFromBytes(data)
		pk, _ = crypto.UnmarshalPrivateKey(data)
		log.Println("load private key from file")
		return pk
	}
	r := rand.Reader
	pk, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	log.Println("create new private key")
	keyByte, err := crypto.MarshalPrivateKey(pk)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(filename, keyByte, 0644)
	return pk
}

func MakeRoutedHost(ctx context.Context, prvKey crypto.PrivKey, listenHost string, listenPort int) (*rhost.RoutedHost, *dht.IpfsDHT) {
	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listenHost, listenPort))

	// Set up stream multiplexer
	// tpt := msmux.NewBlankTransport()
	// tpt.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)
	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	basicHost, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
		libp2p.DefaultTransports,
		//libp2p.Transport(tpt),
		libp2p.NATPortMap(),
		//libp2p.Transport(libp2pquic.NewTransport),
		// Let this host use the DHT to find other hosts
		//libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		//	idht, err = dht.New(ctx, h)
		//	return idht, err
		//}),
		// libp2p.EnableAutoRelay(),
	)

	if err != nil {
		panic(err)
	}

	//==================================================//
	/*

		pid, err := peer.IDFromPrivateKey(prvKey)
		if err != nil {
			panic(err)
		}

		// Create a multiaddress
		addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort))
		if err != nil {
			panic(err)
		}

		// Create a peerstore
		var ps pstore.Peerstore

		// If using secio, we add the keys to the peerstore
		// for this peer ID.
		if secio {
			ps.AddPrivKey(pid, priv)
			//ps.AddPubKey(pid, pub)
		}

		// Set up stream multiplexer
		tpt := msmux.NewBlankTransport()
		tpt.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

		// Create swarm (implements libP2P Network)
		swrm := swarm.NewSwarm(
			context.Background(),
			[]ma.Multiaddr{addr},
			pid,
			ps,
			nil,
			tpt,
			nil,
		)

		netw := (*swarm.Network)(swrm)
		basicHost := bhost.New(netw)
		//==================================================//
	*/
	dstore := dsync.MutexWrap(ds.NewMapDatastore())

	// Make the DHT
	idht := dht.NewDHT(ctx, basicHost, dstore)
	//err = idht.BootstrapWithConfig(ctx, dht.BootstrapConfig{Queries: 1, Period: 5 * time.Second})
	if err = idht.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Make the routed host
	routedHost := rhost.Wrap(basicHost, idht)

	// Bootstrap the host
	err = idht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", routedHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	// addr := routedHost.Addrs()[0]
	addrs := routedHost.Addrs()
	log.Println("I can be reached at:")
	for _, addr := range addrs {
		log.Println(addr.Encapsulate(hostAddr))
	}
	return routedHost, idht

}
