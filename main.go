package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ipsn/go-ipfs/core"
	blocks "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-block-format"
	cid "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-cid"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-floodsub"
	pstore "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-peerstore"
)

const topic = "d3525e66-c98f-4825-a682-7ef1d38f4a72#message"

func main() {
	repo, err := buildRepo()
	if err != nil {
		log.Fatalf("Failed to build repo: %v", err)
	}
	node, err := core.NewNode(context.TODO(), &core.BuildCfg{
		Online:    true,
		Repo:      repo,
		ExtraOpts: map[string]bool{"pubsub": true},
	})
	if err != nil {
		log.Fatalf("Failed to start IPFS node: %v", err)
	}
	go discoverPeers(node)
	pubsub := node.Floodsub
	switch os.Args[1] {
	case "pub":
		pub(pubsub)
	case "sub":
		sub(pubsub)
	}
}

func discoverPeers(n *core.IpfsNode) {
	blk := blocks.NewBlock([]byte("floodsub:" + topic))
	err := n.Blocks.AddBlock(blk)
	if err != nil {
		log.Fatalf("pubsub discovery: %V\n", err)
		return
	}
	connectToPubSubPeers(context.TODO(), n, blk.Cid())
}

func connectToPubSubPeers(ctx context.Context, n *core.IpfsNode, cid cid.Cid) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	provs := n.Routing.FindProvidersAsync(ctx, cid, 10)
	wg := &sync.WaitGroup{}
	for p := range provs {
		wg.Add(1)
		go func(pi pstore.PeerInfo) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			err := n.PeerHost.Connect(ctx, pi)
			if err != nil {
				log.Printf("warn: pubsub discover: %V\n", err)
				return
			}
			log.Printf("connected to pubsub peer: %v\n", pi.ID)
		}(p)
	}

	wg.Wait()
}

func pub(pubsub *floodsub.PubSub) {
	for {
		log.Printf("publishing message\n")
		err := pubsub.Publish(topic, []byte("hello"))
		if err != nil {
			log.Fatalf("Failed to publish on topic: %v", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func sub(pubsub *floodsub.PubSub) {
	s, err := pubsub.Subscribe(topic)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}
	for {
		message, err := s.Next(context.TODO())
		if err != nil {
			log.Fatalf("Failed to get message from topic: %v", err)
		}
		log.Printf("Recieved message: %v\n", message)
	}
}
