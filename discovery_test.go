package discovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	idp "github.com/Rock-liyi/p2pdb-log/identityprovider"
	ks "github.com/Rock-liyi/p2pdb-log/keystore"
	debug "github.com/favframework/debug"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/stretchr/testify/require"
)

func newNode(t *testing.T) host.Host {
	//cm, err := connmgr.NewConnManager(1, 100, connmgr.WithGracePeriod(0))
	//require.NoError(t, err)
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/4001"),
		// We'd like to set the connection manager low water to 0, but
		// that would disable the connection manager.
	//	libp2p.ConnectionManager(cm),
	)
	require.NoError(t, err)
	return h
}

func TestPeeringId(t *testing.T) {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	h1 := newNode(t)
	h2 := newNode(t)
	//ps1 := NewPeeringService(h1)

	debug.Dump(h1.ID().String())
	debug.Dump(h2.ID().String())

}

func TestIdentify(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	datastore := dssync.MutexWrap(NewIdentityDataStore(t))
	keystore, err := ks.NewKeystore(datastore)
	require.NoError(t, err)

	var identities [3]*idp.Identity

	for i, char := range []rune{'A', 'B', 'C'} {
		identity, err := idp.CreateIdentity(ctx, &idp.CreateIdentityOptions{
			Keystore: keystore,
			ID:       fmt.Sprintf("user%c", char),
			Type:     "p2pdb",
		})
		require.NoError(t, err)

		identities[i] = identity
	}
	debug.Dump(identities[1].ID)
}

func TestDiscovery(t *testing.T) {
	//ctx := context.Background()
	Discovery := NewDiscoveryFactory()
	// create a new libp2p Host that listens on a random TCP port
	h, err := Discovery.Create("/ip4/0.0.0.0/tcp/6666")
	debug.Dump(h.ID().String())
	if err != nil {
		panic(err)
	}
	// setup local mDNS discovery
	if err := Discovery.SetupDiscovery(h); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)
}

// func TestPeeringService(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	h1 := newNode(t)
// 	ps1 := NewPeeringService(h1)

// 	h2 := newNode(t)
// 	h3 := newNode(t)
// 	h4 := newNode(t)

// 	// peer 1 -> 2
// 	ps1.AddPeer(peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})
// 	require.Contains(t, ps1.ListPeers(), peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})

// 	// We haven't started so we shouldn't have any peers.
// 	require.Never(t, func() bool {
// 		return len(h1.Network().Peers()) > 0
// 	}, 100*time.Millisecond, 1*time.Second, "expected host 1 to have no peers")

// 	// Use p4 to take up the one slot we have in the connection manager.
// 	for _, h := range []host.Host{h1, h2} {
// 		require.NoError(t, h.Connect(ctx, peer.AddrInfo{ID: h4.ID(), Addrs: h4.Addrs()}))
// 		h.ConnManager().TagPeer(h4.ID(), "sticky-peer", 1000)
// 	}

// 	// Now start.
// 	require.NoError(t, ps1.Start())
// 	// starting twice is fine.
// 	require.NoError(t, ps1.Start())

// 	// We should eventually connect.
// 	t.Logf("waiting for h1 to connect to h2")
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h2.ID()) == network.Connected
// 	}, 30*time.Second, 10*time.Millisecond)

// 	// Now explicitly connect to h3.
// 	t.Logf("waiting for h1's connection to h3 to work")
// 	require.NoError(t, h1.Connect(ctx, peer.AddrInfo{ID: h3.ID(), Addrs: h3.Addrs()}))
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h3.ID()) == network.Connected
// 	}, 30*time.Second, 100*time.Millisecond)

// 	require.Len(t, h1.Network().Peers(), 3)

// 	// force a disconnect
// 	h1.ConnManager().TrimOpenConns(ctx)

// 	// Should disconnect from h3.
// 	t.Logf("waiting for h1's connection to h3 to disconnect")
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h3.ID()) != network.Connected
// 	}, 5*time.Second, 10*time.Millisecond)

// 	// Should remain connected to p2
// 	require.Never(t, func() bool {
// 		return h1.Network().Connectedness(h2.ID()) != network.Connected
// 	}, 5*time.Second, 1*time.Second)

// 	// Now force h2 to disconnect (we have an asymmetric peering).
// 	conns := h2.Network().ConnsToPeer(h1.ID())
// 	require.NotEmpty(t, conns)
// 	h2.ConnManager().TrimOpenConns(ctx)

// 	// All conns to peer should eventually close.
// 	t.Logf("waiting for all connections to close")
// 	for _, c := range conns {
// 		require.Eventually(t, func() bool {
// 			s, err := c.NewStream(context.Background())
// 			if s != nil {
// 				_ = s.Reset()
// 			}
// 			return err != nil
// 		}, 5*time.Second, 10*time.Millisecond)
// 	}

// 	// Should eventually re-connect.
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h2.ID()) == network.Connected
// 	}, 30*time.Second, 1*time.Second)

// 	// Unprotect 2 from 1.
// 	ps1.RemovePeer(h2.ID())
// 	require.NotContains(t, ps1.ListPeers(), peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})

// 	// Trim connections.
// 	h1.ConnManager().TrimOpenConns(ctx)

// 	// Should disconnect
// 	t.Logf("waiting for h1 to disconnect from h2")
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h2.ID()) != network.Connected
// 	}, 5*time.Second, 10*time.Millisecond)

// 	// Should never reconnect.
// 	t.Logf("ensuring h1 is not connected to h2 again")
// 	require.Never(t, func() bool {
// 		return h1.Network().Connectedness(h2.ID()) == network.Connected
// 	}, 20*time.Second, 1*time.Second)

// 	// Until added back
// 	ps1.AddPeer(peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})
// 	require.Contains(t, ps1.ListPeers(), peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})
// 	ps1.AddPeer(peer.AddrInfo{ID: h3.ID(), Addrs: h3.Addrs()})
// 	require.Contains(t, ps1.ListPeers(), peer.AddrInfo{ID: h3.ID(), Addrs: h3.Addrs()})
// 	t.Logf("wait for h1 to connect to h2 and h3 again")
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h2.ID()) == network.Connected
// 	}, 30*time.Second, 1*time.Second)
// 	require.Eventually(t, func() bool {
// 		return h1.Network().Connectedness(h3.ID()) == network.Connected
// 	}, 30*time.Second, 1*time.Second)

// 	// Should be able to repeatedly stop.
// 	require.NoError(t, ps1.Stop())
// 	require.NoError(t, ps1.Stop())

// 	// Adding and removing should work after stopping.
// 	ps1.AddPeer(peer.AddrInfo{ID: h4.ID(), Addrs: h4.Addrs()})
// 	require.Contains(t, ps1.ListPeers(), peer.AddrInfo{ID: h4.ID(), Addrs: h4.Addrs()})
// 	ps1.RemovePeer(h2.ID())
// 	require.NotContains(t, ps1.ListPeers(), peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})
// }

// func TestNextBackoff(t *testing.T) {
// 	minMaxBackoff := (100 - maxBackoffJitter) / 100 * maxBackoff
// 	for x := 0; x < 1000; x++ {
// 		ph := peerHandler{nextDelay: time.Second}
// 		for min, max := time.Second*3/2, time.Second*5/2; min < minMaxBackoff; min, max = min*3/2, max*5/2 {
// 			b := ph.nextBackoff()
// 			if b > max || b < min {
// 				t.Errorf("expected backoff %s to be between %s and %s", b, min, max)
// 			}
// 		}
// 		for i := 0; i < 100; i++ {
// 			b := ph.nextBackoff()
// 			if b < minMaxBackoff || b > maxBackoff {
// 				t.Fatal("failed to stay within max bounds")
// 			}
// 		}
// 	}
// }
