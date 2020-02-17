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

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"

	//"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
)

var idht *dht.IpfsDHT

//var idht *dht.IpfsDHT

func main() {
	cfg := parseFlags()
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
		//libp2p.EnableAutoRelay(),
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
	/*
		// connect to the chosen ipfs nodes
		err = bootstrapConnect(ctx, routedHost, bootstrapPeers)
		if err != nil {
			panic(err)
		}*/

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

	//log.Printf("Now run \"./routed-echo -l %d -d %s%s\" on a different terminal\n", listenPort+1, routedHost.ID().Pretty(), globalFlag)

	go printTable(ctx, routedHost.ID())
	//go printNetworkPeers(routedHost)
	routedHost.SetStreamHandler(protocol.ID(cfg.ProtocolID), handleStream)
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
	fmt.Println("-----RemoteAddr--------")
	fmt.Println(stream.Conn().RemoteMultiaddr().String())
	fmt.Println(stream.Conn().RemotePeer().String())
	fmt.Println("-----My--------")
	fmt.Println(stream.Conn().LocalPeer().String())
	fmt.Println(stream.Conn().LocalMultiaddr().String())
	fmt.Println(idht.RoutingTable().ListPeers())
}

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

//Initialize the MDNS service
func initMDNS(ctx context.Context, peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	// An hour might be a long long period in practical applications. But this is fine for us
	ser, err := discovery.NewMdnsService(ctx, peerhost, time.Hour, rendezvous)
	if err != nil {
		panic(err)
	}

	//register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser.RegisterNotifee(n)
	return n.PeerChan
}

func printTable(ctx context.Context, p peer.ID) {
	for {
		//idht.Update(ctx, p)
		//idht.BootstrapSelf(ctx)
		//fmt.Println(idht.Context())
		fmt.Print(len(idht.RoutingTable().ListPeers()))
		fmt.Println(idht.RoutingTable().ListPeers())
		//fmt.Println(idht.FindPeer(ctx, "QmWY9uYqDBCbf29TCJWfnhrBVEteQdXzSuG3LKop5nHbHS"))
		time.Sleep(5 * time.Second)
	}
}

/*
func printNetworkPeers(rountedHost *rhost.RoutedHost) {
	for {
		fmt.Println(rountedHost.Network().Peers())
		fmt.Println(rountedHost.Network().Peerstore().Peers())
		time.Sleep(5 * time.Second)
	}
}*/

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
