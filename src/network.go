package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Network messages
// 2 bit message description:
// 00: PING
// 01: STORE
// 10: FIND_NODE
// 11: FIND_VALUE
// Followed by a 20 byte ID of needed and data for STORE command

// Golang doesn't have enums, this the closest alternative I could find
const (
	PING byte = iota // 0
	STORE byte = iota // 1
	FIND_NODE byte = iota // 2
	FIND_VALUE byte = iota // 3
)

type Network struct {
	k int
	alpha int
	localNode *Node
}

func NewNetwork(node *Node) Network {
	return Network{20, 3, node}
}

func Listen(ip string, port int) {
	net.Listen("udp", ip + string(port))
}

func (network *Network) SendPingMessage(contact *Contact) {
	reply := make([]byte, 256)

	conn, err := net.Dial("udp", contact.Address + ":5001")
	fmt.Println("Connection established! Sending ping msg.")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
		return
	}

	start := time.Now()
	conn.Write([]byte("This is a ping msg!" + "\n"))
	conn.Read(reply)
	duration := time.Since(start)

	fReply := strings.Split(string(reply), "\n")
	if fReply[0] == "Ack!" {
		fmt.Println("Pinging node " + contact.ID.String() + " took " + strconv.FormatInt(duration.Milliseconds(),
			10) + " ms")
	} else {
		fmt.Println("Received unrecognized response from node " + contact.ID.String() + " when pinged")
	}
}

func (network *Network) NodeLookup(lookupID *KademliaID) []Contact {
	// TODO
	// This will be the central node lookup algorithm that can be used to find nodes or data depending on
	// the lookup ID. It will always try to locate the K closest nodes in the network and starts by sending
	// udp messages and recursively locates nodes that are closer until no more nodes can be found. Each node
	// will then receive these messages and search through their own routing table

	myID := network.localNode.routingTable.me.ID
	network.localNode.routingTable.FindClosestContacts(myID, network.k)

	network.localNode.routingTable.getBucketIndex(lookupID)

	return nil
}

/* This is not needed?
func (network *Network) SendFindContactMessage(contact *Contact) {
	conn, err := net.Dial("udp", contact.Address + ":5001")
}
 */

func (network *Network) SendFindDataMessage(hash KademliaID) {
	// Prepare FIND_VALUE RPC
	findMessage := make([]byte, 1)
	findMessage[0] = FIND_VALUE
	findMessage = append(findMessage, hash[:]...)

	var nodes = network.NodeLookup(hash) // Get ALL nodes that are closest to the hash value
	for _,contact := range nodes { // What type of syntax is this??
		if network.localNode.routingTable.me.ID == contact.ID {
			// No need to send a network request. Send the RPC directly to the local node thread.
		} else {
			// TODO Send a FIND_VALUE RPC
		}
	}
}

func (network *Network) SendStoreMessage(hash KademliaID,data []byte) {
	// Prepare STORE RPC
	storeMessage := make([]byte, 1)
	storeMessage[0] = STORE
	storeMessage = append(storeMessage, hash[:]...)
	storeMessage = append(storeMessage, data...)

	var nodes = network.NodeLookup(hash) // Get ALL nodes that are closest to the hash value
	for _,contact := range nodes { // What type of syntax is this??
		if network.localNode.routingTable.me.ID == contact.ID {
			// No need to send a network request. Send the RPC directly to the local node thread.
		} else {
			// TODO Send a STORE RPC
		}
	}
}
