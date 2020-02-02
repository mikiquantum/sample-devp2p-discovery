package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestReverse(t *testing.T) {
	input := "THISISATEST"
	expected := "TSETASISIHT"
	output := reverse(input)
	if output != expected {
		t.Error("Expected", expected)
	}
}

func TestDiscoveryAndMessageExchange(t *testing.T) {
	// BootNode setup
	bootLogger := log.New("name", "boot")
	logHandler := log.StreamHandler(io.Writer(os.Stdout), log.TerminalFormat(true))
	bootLogger.SetHandler(log.LvlFilterHandler(log.LvlInfo, logHandler))
	discNode, err := StartBootnode(":30301", newkey(), bootLogger)
	if err != nil {
		t.Error(err)
	}

	bnode, err := enode.ParseV4(fmt.Sprintf("%s@%s",  discNode.Self().URLv4(), "127.0.0.1:0?discport=30301"))
	if err != nil {
		t.Error(err)
	}

	blist := []*enode.Node{bnode}
	fmt.Println("Started Bootnode with ID", discNode.Self().ID(), "giving some time to init...")
	time.Sleep(5 * time.Second)

	// Node1 Setup
	n1Logger := log.New("name", "n1")
	n1Logger.SetHandler(log.LvlFilterHandler(log.LvlInfo, logHandler))
	rh1 := new(ReverseHandler)
	rp1 := NewReverseProtocol(rh1)
	n1, err := StartNodeServer("n1", n1Logger,30302, blist, []p2p.Protocol{rp1})
	if err != nil {
		t.Errorf("Could not start server: %v", err)
	}
	defer n1.Stop()
	fmt.Println("Started Node 1", n1.NodeInfo().ID)

	// Node2 Setup
	n2Logger := log.New("name", "n2")
	n2Logger.SetHandler(log.LvlFilterHandler(log.LvlInfo, logHandler))
	rh2 := new(ReverseHandler)
	rp2 := NewReverseProtocol(rh2)
	n2, err := StartNodeServer("n2", n2Logger, 30303, blist, []p2p.Protocol{rp2})
	if err != nil {
		t.Errorf("Could not start server: %v", err)
	}
	defer n2.Stop()
	fmt.Println("Started Node 2", n2.NodeInfo().ID)
	err = waitForPeer(n2, 20, n1.NodeInfo().ID)
	if err != nil {
		t.Error(err)
	}
	err = waitForPeer(n1, 20, n2.NodeInfo().ID)
	if err != nil {
		t.Error(err)
	}

	// Make use of each nodes handler instance to allow >5 msg iterations before shutting down
	for {
		if rh1.totalCounter > 5  && rh2.totalCounter > 5 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func waitForPeer(n *p2p.Server, maxRetries int, reqPeerID string) error {
	count := 0
	loop:
	for {
		if count == maxRetries {
			return errors.New("timeout discovering peers")
		}
		for _, p := range n.Peers() {
			if p.ID().String() == reqPeerID {
				break loop
			}
		}
		count++
		time.Sleep(5 * time.Second)
	}
	return nil
}
