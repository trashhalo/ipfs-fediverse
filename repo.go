package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	ds "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-datastore"
	dsync "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-datastore/sync"
	cfg "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-config"
	ci "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-crypto"
	peer "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-peer"
	repo "github.com/ipsn/go-ipfs/repo"
	"github.com/phayes/freeport"
)

func buildRepo() (repo.Repo, error) {
	var d ds.Datastore
	d = ds.NewMapDatastore()
	bootstrapPeers, err := cfg.DefaultBootstrapPeers()
	if err != nil {
		return nil, err
	}
	c := cfg.Config{
		Bootstrap: cfg.BootstrapPeerStrings(bootstrapPeers),
		Pubsub:    cfg.PubsubConfig{Router: "gossipsub"},
		Discovery: cfg.Discovery{
			MDNS: cfg.MDNS{
				Enabled:  true,
				Interval: 10,
			},
		},
		Routing: cfg.Routing{
			Type: "dht",
		},
		Ipns: cfg.Ipns{
			ResolveCacheSize: 128,
		},
		Reprovider: cfg.Reprovider{
			Interval: "12h",
			Strategy: "all",
		},
		Swarm: cfg.SwarmConfig{
			ConnMgr: cfg.ConnMgr{
				LowWater:    cfg.DefaultConnMgrLowWater,
				HighWater:   cfg.DefaultConnMgrHighWater,
				GracePeriod: cfg.DefaultConnMgrGracePeriod.String(),
				Type:        "basic",
			},
		},
	}
	priv, pub, err := ci.GenerateKeyPairWithReader(ci.RSA, 1024, rand.Reader)
	if err != nil {
		return nil, err
	}

	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	privkeyb, err := priv.Bytes()
	if err != nil {
		return nil, err
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	c.Addresses.Swarm = []string{fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)}
	c.Identity.PeerID = pid.Pretty()
	c.Identity.PrivKey = base64.StdEncoding.EncodeToString(privkeyb)

	return &repo.Mock{
		D: dsync.MutexWrap(d),
		C: c,
	}, nil
}
