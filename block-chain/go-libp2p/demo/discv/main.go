package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/libp2p/go-libp2p-core/protocol"
	net "github.com/libp2p/go-libp2p-net"
	"google.golang.org/grpc"

	//"github.com/ipfs/go-log"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/demo/discv/echosvc"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
	// "google.golang.org/grpc"
)

var idht *dht.IpfsDHT
var routedHost *rhost.RoutedHost

//var idht *dht.IpfsDHT
const proto string = "/grpc/0.0.1"

var cfg = parseFlags()

// Echoer implements the EchoService.
type Echoer struct {
	PeerID peer.ID
}

// Echo asks a node to respond with a message.
func (e *Echoer) Echo(ctx context.Context, req *echosvc.EchoRequest) (*echosvc.EchoReply, error) {
	fmt.Println("handling msg")
	return &echosvc.EchoReply{
		Message: req.GetMessage(),
		PeerId:  e.PeerID.Pretty(),
	}, nil
}

func main() {
	// create a background context (i.e. one that never cancels)
	ctx := context.Background()

	// Creates a new RSA key
	//prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	prvKey := GetPrivateKey(cfg.Name)

	routedHost, idht = MakeRoutedHost(ctx, prvKey, cfg.listenHost, cfg.listenPort)

	// go printTable(ctx, routedHost.ID())
	// go broadCastNetworkPeers(ctx, routedHost)
	//go printNetworkPeers(routedHost)

	// routedHost.SetStreamHandler(protocol.ID(proto), handleStream)
	var peer peer.AddrInfo
	if cfg.bootstrap == "" {
		// peerChan := initMDNS(ctx, routedHost, proto)
		// peer = <-peerChan // will block untill we discover a peer
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
		if err := routedHost.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
		}
	}

	fmt.Println("Found peer:", peer, ", connecting")

	//========================Codes for Discovery Ends =================================//
	//========================Codes for GRPC Begins ===================================//
	// Set the grpc protocol handler on it
	grpcProto := p2pgrpc.NewGRPCProtocol(context.Background(), routedHost)

	// Register our echoer GRPC service.
	// go RegistrateService(grpcProto)
	echosvc.RegisterEchoServiceServer(grpcProto.GetGRPCServer(), &Echoer{PeerID: routedHost.ID()})

	go BroadCastThroughGrpc(grpcProto)
	/*
		if *target == "" {
			log.Println("listening for connections")
			select {} // hang forever
		}*/
	/**** This is where the listener code ends ****/

	//========================Codes for GRPC Ends =========================================//

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
			hello := "from " + cfg.Name + "\n"
			_, err = stream.Write([]byte(hello))
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func RegistrateService(grpcProto *p2pgrpc.GRPCProtocol) {
	// time.Sleep(10 * time.Second)
	echosvc.RegisterEchoServiceServer(grpcProto.GetGRPCServer(), &Echoer{PeerID: routedHost.ID()})
	select {}
}

func BroadCastThroughGrpc(grpcProto *p2pgrpc.GRPCProtocol) {
	for {
		time.Sleep(20 * time.Second)
		fmt.Println("sending msg to following peers:")
		fmt.Println(idht.RoutingTable().ListPeers())
		for _, peerid := range idht.RoutingTable().ListPeers() {
			// make a new stream from host B to host A
			log.Println("dialing via grpc")
			grpcConn, err := grpcProto.Dial(context.Background(), peerid, grpc.WithInsecure())
			if err != nil {
				log.Fatalln(err)
			}
			// create our service client
			echoClient := echosvc.NewEchoServiceClient(grpcConn)
			echoMsg := "send from: " + cfg.Name + "\t"
			echoReply, err := echoClient.Echo(context.Background(), &echosvc.EchoRequest{Message: echoMsg})
			if err != nil {
				log.Fatalln(err)
			}

			log.Println("read reply:")
			err = (&jsonpb.Marshaler{EmitDefaults: true, Indent: "\t"}).
				Marshal(os.Stdout, echoReply)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println()
		}
	}
}

// doEcho reads a line of data a stream and writes it back
func doEcho(s net.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return err
}
