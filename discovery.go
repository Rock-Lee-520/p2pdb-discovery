package discovery

import (
	"context"
	"crypto/rand"
	"github.com/Rock-liyi/p2pdb/infrastructure/util/log"
	"github.com/libp2p/go-libp2p"
	libp2pConnmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	libp2pTls "github.com/libp2p/go-libp2p/p2p/security/tls"
	libp2pTcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

const LISTEN_ADDRESS_STRINGS = "/ip4/0.0.0.0/tcp/0"

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// code form https://github.com/IceFireDB/IceFireDB/tree/main/IceFireDB-PubSub/pkg/p2p
// Initialize network in ipfs
func InitDiscovery(ctx context.Context, address string) (host.Host, *dht.IpfsDHT) {
	if address == "" {
		address = LISTEN_ADDRESS_STRINGS
	}

	// Setup Identity
	prvkey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		log.Debug("p2pdb-discovery Setup Identity fail:", err.Error())
	}
	identity := libp2p.Identity(prvkey)

	// Setup TLS
	tls, err := libp2pTls.New(prvkey)
	security := libp2p.Security(libp2pTls.ID, tls)
	transport := libp2p.Transport(libp2pTcp.NewTCPTransport)
	if err != nil {
		log.Debug("p2pdb-discovery Setup TLS fail:", err.Error())
	}

	// SetUp listener address
	muladdr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		log.Debug("p2pdb-discovery SetUp listener address fail:", err.Error())
	}
	listen := libp2p.ListenAddrs(muladdr)
	// Set up the stream multiplexer and connection manager options
	muxer := libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport)
	connmgr, err := libp2pConnmgr.NewConnManager(100, 400)
	if err != nil {
		log.Debug("p2pdb-discovery NewConnManager fail:", err.Error())
	}
	conn := libp2p.ConnectionManager(connmgr)

	// Setup NAT
	nat := libp2p.NATPortMap()
	relay := libp2p.EnableAutoRelay()

	// Declare a KadDHT
	var kaddht *dht.IpfsDHT
	// Setup a routing configuration with the KadDHT
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		kaddht = setupKadDHT(ctx, h)
		return kaddht, err
	})

	opts := libp2p.ChainOptions(identity, listen, security, transport, muxer, conn, nat, routing, relay)

	// Construct a new libP2P host with the created options
	libhost, err := libp2p.New(opts)
	if err != nil {
		log.Debug("p2pdb-discovery new libp2p fail:", err.Error())
	}
	return libhost, kaddht
}

// code form https://github.com/IceFireDB/IceFireDB/tree/main/IceFireDB-PubSub/pkg/p2p
// A function that generates a Kademlia DHT object and returns it
func setupKadDHT(ctx context.Context, nodehost host.Host) *dht.IpfsDHT {
	// Create DHT server mode option
	dhtmode := dht.Mode(dht.ModeServer)
	// Rertieve the list of boostrap peer addresses
	bootstrappeers := dht.GetDefaultBootstrapPeerAddrInfos()
	// Create the DHT bootstrap peers option
	dhtpeers := dht.BootstrapPeers(bootstrappeers...)

	// Trace log
	log.Debug("Generated DHT Configuration.")

	// Start a Kademlia DHT on the host in server mode
	kaddht, err := dht.New(ctx, nodehost, dhtmode, dhtpeers)
	// Handle any potential error
	if err != nil {
		log.Debug("error", err.Error(), "Failed to Create the Kademlia DHT!")
	}

	// Return the KadDHT
	return kaddht
}

// code form https://github.com/IceFireDB/IceFireDB/tree/main/IceFireDB-PubSub/pkg/p2p
// A function that bootstraps a given Kademlia DHT to satisfy the IPFS router
// interface and connects to all the bootstrap peers provided by libp2p
func BootstrapDHT(ctx context.Context, nodehost host.Host, kaddht *dht.IpfsDHT) {
	// Bootstrap the DHT to satisfy the IPFS Router interface
	if err := kaddht.Bootstrap(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatalln("Failed to Bootstrap the Kademlia!")
	}

	// Trace log
	logrus.Traceln("Set the Kademlia DHT into Bootstrap Mode.")

	// Declare a WaitGroup
	var wg sync.WaitGroup
	// Declare counters for the number of bootstrap peers
	var connectedbootpeers int32
	var totalbootpeers int32

	// Iterate over the default bootstrap peers provided by libp2p
	for _, peeraddr := range dht.DefaultBootstrapPeers {
		// Retrieve the peer address information
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peeraddr)

		// Incremenent waitgroup counter
		wg.Add(1)
		totalbootpeers++
		// Start a goroutine to connect to each bootstrap peer
		go func() {
			// Defer the waitgroup decrement
			defer wg.Done()
			// Attempt to connect to the bootstrap peer
			if err := nodehost.Connect(ctx, *peerinfo); err == nil {
				// Increment the connected bootstrap peer count
				atomic.AddInt32(&connectedbootpeers, 1)
			}
		}()
	}

	// Wait for the waitgroup to complete
	wg.Wait()

	// Log the number of bootstrap peers connected
	log.Debug("Connected to %d out of %d Bootstrap Peers.", connectedbootpeers, totalbootpeers)
}

func NewRoutingDiscovery(router routing.ContentRouting) *discovery.RoutingDiscovery {
	return &discovery.RoutingDiscovery{router}
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	var localPeerId = n.h.ID()
	if pi.ID.Pretty() != localPeerId.String() {
		log.Info("discovered new remote peer id %s\n", pi.ID.Pretty())
	}

	err := n.h.Connect(context.Background(), pi)

	if pi.ID.Pretty() != localPeerId.String() && err != nil {
		log.Info("error connecting to remote peer id %s: %s\n", pi.ID.Pretty(), err)
	}
}
