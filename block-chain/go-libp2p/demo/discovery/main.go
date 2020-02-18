package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p-core/protocol"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"

	//"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	crypto "github.com/libp2p/go-libp2p-crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
)

var idht *dht.IpfsDHT

//var idht *dht.IpfsDHT
const proto string = "HBC/O.0.0"

var cfg = parseFlags()

func main() {
	//cfg := parseFlags()
	// create a background context (i.e. one that never cancels)
	ctx := context.Background()

	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
	// Creates a new RSA key pair for this host.
	r := rand.Reader
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	basicHost, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
		libp2p.DefaultTransports,
		libp2p.NATPortMap(),
		//libp2p.Transport(libp2pquic.NewTransport),
		// Let this host use the DHT to find other hosts
		/*libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),*/
		// libp2p.EnableAutoRelay(),
	)

	if err != nil {
		panic(err)
	}

	dstore := dsync.MutexWrap(ds.NewMapDatastore())

	// Make the DHT
	idht = dht.NewDHT(ctx, basicHost, dstore)
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

	// go printTable(ctx, routedHost.ID())
	go broadCastNetworkPeers(ctx, routedHost)
	//go printNetworkPeers(routedHost)
	routedHost.SetStreamHandler(protocol.ID(proto), handleStream)
	var peer peer.AddrInfo
	if cfg.bootstrap == "" {
		peerChan := initMDNS(ctx, routedHost, "meetme")
		peer = <-peerChan // will block untill we discover a peer
	} else {
		addrs, err := multiaddr.NewMultiaddr(cfg.bootstrap)
		if err != nil {
			panic(err)
		}
		peerA, err := peerstore.InfoFromP2pAddr(addrs)
		peer = *peerA
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Found peer:", peer, ", connecting")

	if err := routedHost.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:", err)
	}
	//routingDiscovery := discovery2.NewRoutingDiscovery(idht)

	fmt.Println("----------------------------")
	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")
	// shut the node down
	if err := routedHost.Close(); err != nil {
		panic(err)
	}
}

func handleStream(stream network.Stream) {
	fmt.Println("handling msg")
	buf := bufio.NewReader(stream)
	str, err := buf.ReadString('\n')
	//str, err := ioutil.ReadAll(stream)
	if err != nil {
		panic(err)
	}
	fmt.Println("Received:" + str)
}

func printTable(ctx context.Context, p peer.ID) {
	for {
		fmt.Print(len(idht.RoutingTable().ListPeers()))
		fmt.Println(idht.RoutingTable().ListPeers())
		time.Sleep(30 * time.Second)
	}
}

func broadCastNetworkPeers(ctx context.Context, routedHost *rhost.RoutedHost) {
	for {
		time.Sleep(30 * time.Second)
		fmt.Println("sending msg to following peers:")
		fmt.Println(idht.RoutingTable().ListPeers())
		for _, peer := range idht.RoutingTable().ListPeers() {
			stream, err := routedHost.NewStream(context.Background(), peer, protocol.ID(proto))
			hello := "from " + cfg.id + "\n"
			_, err = stream.Write([]byte(hello))
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
