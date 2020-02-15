package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/qjpcpu/ethereum/key"
	"github.com/qjpcpu/log"
)

type MessageType = uint64

const (
	MT_HelloWorld MessageType = iota
	MT_FooBar
)

type Message string

func FooBarProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    "FooBarProtocol",
		Version: 1,
		Length:  2,
		Run:     msgHandler,
	}
}

var (
	nodeName  string
	uid       string
	port      string
	bootnode  string
	peersList string
)

func init() {
	peersList = "peers"
	rand.Seed(time.Now().UTC().UnixNano())
	uid = strconv.FormatUint(rand.Uint64(), 10)
	flag.StringVar(&nodeName, "name", "", "node name")
	flag.StringVar(&port, "port", "", "listen port")
	flag.StringVar(&bootnode, "bootstrap", "", "bootstrap node")
}

func bootstrapNodes() []*enode.Node {
	var nodes []*enode.Node
	if bootnode != "" {
		log.Infof("bootstrap nodes:%+v", bootnode)
		//nodes = append(nodes, discv5.MustParseNode(bootnode))
		nodes = append(nodes, enode.MustParse(bootnode))
	}
	return nodes
}

func parseArgs() {
	flag.Parse()
	if port == "" {
		log.Error("no port")
		os.Exit(1)
	}
	if nodeName == "" {
		log.Error("no node name")
		os.Exit(1)
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
}

func getPrivateKey() *ecdsa.PrivateKey {
	os.MkdirAll(nodeName, 0777)
	filename := nodeName + "/private-key"
	var pk *ecdsa.PrivateKey
	for loop := true; loop; loop = false {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			break
		}
		pk = key.PrivateKeyFromBytes(data)
		log.Info("load private key from file")
		return pk
	}
	pk, _ = crypto.GenerateKey()
	log.Info("create new private key")
	ioutil.WriteFile(filename, key.PrivateKeyToBytes(pk), 0644)
	return pk
}

var srv p2p.Server
var peers []*enode.Node

func main() {
	parseArgs()
	nodekey := getPrivateKey()
	srv = p2p.Server{
		Config: p2p.Config{
			MaxPeers:       10,
			PrivateKey:     nodekey,
			Name:           uid,
			ListenAddr:     port,
			Protocols:      []p2p.Protocol{FooBarProtocol()},
			NAT:            nat.Any(),
			BootstrapNodes: bootstrapNodes(),
			// BootstrapNodes: peers,
		},
	}
	if err := srv.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Info("started..", srv.NodeInfo().Enode)
	ch := make(chan *p2p.PeerEvent)
	srv.SubscribeEvents(ch)
	// srv.AddPeer(enode.MustParse("enode://09d05fdfc30f024c75b4d7e304f9929616a1a525bcce90188fa5981596c732299715f3cef108ef7de9ee001643cafe2492b6036464050b5a35679f7cd55ccea7@127.0.0.1:30330"))
	for {
		select {
		case <-time.After(60 * time.Second):
			log.Infof("connected %d nodes", srv.PeerCount())
		case pe := <-ch:
			log.Info("PE", pe.Type, pe.Protocol, pe.Peer.String())
		}
	}
}

func msgHandler(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	// send msg
	log.Infof("new peer connected:%v", peer.String())
	p2p.SendItems(ws, MT_HelloWorld, srv.NodeInfo().Name+":welcome "+peer.Name())
	p2p.SendItems(ws, MT_FooBar, srv.NodeInfo().Enode)
	for {
		msg, err := ws.ReadMsg()
		if err != nil {
			log.Warningf("peer %s disconnected", peer.Name())
			return err
		}

		var myMessage [1]Message
		err = msg.Decode(&myMessage)
		if err != nil {
			log.Errorf("decode msg err:%v", err)
			// handle decode error
			continue
		}

		log.Info("code:", msg.Code, "receiver at:", msg.ReceivedAt, "msg:", myMessage[0])
		/*
			switch myMessage[0] {
			case "foo":
				err := p2p.SendItems(ws, MT_FooBar, "bar")
				if err != nil {
					log.Errorf("send bar error:%v", err)
					return err
				}
			default:
				//msg2, _ := strconv.ParseUint(srv.NodeInfo().Enode, 0, 64)
				err := p2p.SendItems(ws, MT_FooBar, srv.NodeInfo().Enode)
				if err != nil {
					log.Errorf("send bar error:%v", err)
					return err
				}
				log.Info("recv:", myMessage)
			}

				fmt.Println("-----------------------")
				log.Info("code:", msg.Code, "receiver at:", msg.ReceivedAt, "msg:", myMessage)
				fmt.Println("-----------------------")*/
	}
}
