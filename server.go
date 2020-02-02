package main

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

const messageId = 0

// Message defines the structure that is RLP-serialized
type Message struct {
	Seq uint
	Val string
}

//func (m *Message) EncodeRLP(w io.Writer) error {
//	err := rlp.Encode(w, m.Seq)
//	if err != nil {
//		return err
//	}
//	return rlp.Encode(w, m.Val)
//}
//
//func (m *Message) DecodeRLP(s *rlp.Stream) error {
//	err := s.Decode(&m.Seq)
//	if err != nil {
//		return err
//	}
//	return s.Decode(&m.Val)
//}

// NewReverseProtocol returns a Protocol that reverses a string
func NewReverseProtocol(reverseHandler *ReverseHandler) p2p.Protocol {
	return p2p.Protocol{
		Name:    "reverse",
		Version: 1,
		Length:  1,
		Run:     reverseHandler.handlerFunc,
	}
}

// ReverseHandler holds certain counter of the number of rounds
type ReverseHandler struct {
	totalCounter int
}

func (m *ReverseHandler) handlerFunc(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	rand.Seed(time.Now().UnixNano()) // Random seed init per peer

	idx := uint(0)
	go func() {
		for range time.Tick(2 * time.Second) {
			msgg := &Message{Seq: idx + 1, Val: randomString(5)}
			peer.Log().Info("Sending Request", "seq", msgg.Seq, "value", msgg.Val)
			err := p2p.Send(ws, messageId, msgg)
			if err != nil {
				peer.Log().Error("error sending", err)
				return
			}
		}
	}()

	for {
		msg, err := ws.ReadMsg()
		if err != nil {
			return err
		}
		var myMessage Message
		err = msg.Decode(&myMessage)
		if err != nil {
			return err
		}
		if myMessage.Seq == idx { // Receive response for current sequence, closing loop for that sequence
			peer.Log().Info("Received Response", "seq", myMessage.Seq, "value", myMessage.Val)
			m.totalCounter++
		} else {
			peer.Log().Info("Received Request", "seq", myMessage.Seq, "value", myMessage.Val)
			nm := &Message{Seq: myMessage.Seq, Val: reverse(myMessage.Val)}
			peer.Log().Info("Sending Response", "seq", nm.Seq, "value", nm.Val)
			err = p2p.Send(ws, messageId, nm)
			if err != nil {
				return err
			}
			idx = myMessage.Seq
		}
	}
}

// StartNodeServer starts a p2p server and returns its instance
func StartNodeServer(name string, logger log.Logger, port int, bootnodes []*enode.Node, protocols []p2p.Protocol) (*p2p.Server, error) {
	config := p2p.Config{
		Name:       name,
		MaxPeers:   10,
		ListenAddr: fmt.Sprintf("127.0.0.1:%d", port),
		PrivateKey: newkey(),
		BootstrapNodes: bootnodes,
		Logger: logger,
		Protocols: protocols,
	}

	server := &p2p.Server{
		Config:       config,
	}
	if err := server.Start(); err != nil {
		return nil, err
	}
	return server, nil
}

// StartBootnode starts a UDPv4 discovery listener and returns instance
func StartBootnode(bootAddr string, key *ecdsa.PrivateKey, log log.Logger) (*discover.UDPv4, error) {
	addr, err := net.ResolveUDPAddr("udp", bootAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	// In memory routing table
	db, err := enode.OpenDB("")
	if err != nil {
		return nil, err
	}

	ln := enode.NewLocalNode(db, key)
	cfg := discover.Config{
		PrivateKey:  key,
		Log: log,
	}

	udpV4, err := discover.ListenUDP(conn, ln, cfg)
	if err != nil {
		return nil, err
	}

	return udpV4, nil
}

// Tools

func newkey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(length int) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(randomInt(65, 90)) //Capital letters only
	}
	return string(bytes)
}
