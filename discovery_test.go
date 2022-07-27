package discovery

import (
	"context"
	"fmt"
	"testing"

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
