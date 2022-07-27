package test

import (
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p"
)

//get addresses from the server	 default
func TestAddress(t *testing.T) {
	// start a libp2p node with default settings
	node, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	// print the node's listening addresses
	fmt.Println("Listen addresses:", node.Addrs())
	fmt.Println("Listen id:", node.ID())

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}

//get  id by setting addrstring
func TestPeerId(t *testing.T) {
	// start a libp2p node with default settings
	node, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/2000"))
	if err != nil {
		panic(err)
	}

	// print the node's listening addresses
	fmt.Println("Listen addresses:", node.Addrs())
	fmt.Println("Listen id:", node.ID())

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
