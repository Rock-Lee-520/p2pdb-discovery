package discovery

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "p2pdb-example"

type Discovery interface {
	// create a new libp2p Host that listens on a random TCP port
	//	Connect() (Host string)
	Create(host string) (host.Host, error)
	SetupDiscovery(host host.Host) error
}

type DiscoveryFactory struct {
	h host.Host
}

const LISTEN_ADDRESS_STRINGS = "/ip4/0.0.0.0/tcp/0"

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

func (d *DiscoveryFactory) GetLocalPeerId() peer.ID {
	return d.h.ID()
}

func (d *DiscoveryFactory) SetLocalPeerId(id peer.ID) {

}

func NewDiscoveryFactory() *DiscoveryFactory {
	return &DiscoveryFactory{}
}

func CreatePeerKey(key string) (string, error) {
	return key, nil
}

func GetPublicKey(key string) (string, error) {
	return key, nil
}

func GetPrivateKey(key string) (string, error) {
	return key, nil
}

// func (d *DiscoveryFactory) Start(publicKey string) (host.Host, error) {

// 	h, err := libp2p.New(libp2p.ListenAddrStrings(LISTEN_ADDRESS_STRINGS))

// 	return h, err
// }

// func (d *DiscoveryFactory) Connect() (Host string)  (Discovery, error){
// 	return d,nil
// }

func (d *DiscoveryFactory) Create(host string) (host.Host, error) {
	if host == "" {
		host = LISTEN_ADDRESS_STRINGS
	}
	var options = libp2p.ListenAddrStrings(host)

	h, err := libp2p.New(options)
	d.h = h
	return h, err
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func (d *DiscoveryFactory) SetupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	log.Printf("local peer id is %s\n", h.ID())
	return s.Start()
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	var localPeerId = n.h.ID()
	if pi.ID.Pretty() != localPeerId.String() {
		log.Printf("discovered new remote peer id %s\n", pi.ID.Pretty())
	}

	err := n.h.Connect(context.Background(), pi)

	if pi.ID.Pretty() != localPeerId.String() && err != nil {
		log.Printf("error connecting to remote peer id %s: %s\n", pi.ID.Pretty(), err)
	}
}
