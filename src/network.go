package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"encoding/json"
)

// Network messages
// 2 bit message description:
// 000: PING
// 001: STORE
// 010: FIND_NODE
// 011: FIND_VALUE
// 100: REPLY (???)
// Followed by a 20 byte ID of needed and data for STORE command
// 2 bit + 20 byte +

// Golang doesn't have enums, this the closest alternative I could find
const (
	PING byte = 0 // 0
	STORE byte = 1 // 1
	FIND_NODE byte = 2 // 2
	FIND_VALUE byte = 3 // 3
	STORE_DATA_SIZE int = iota
	BUCKET_DATA_SIZE int = iota
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

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
		return
	} else {
		fmt.Println("Connection established! Sending ping msg.")
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

func (network *Network) NodeLookup(lookupID KademliaID) []Contact {
	// This will be the central node lookup algorithm that can be used to find nodes or data depending on
	// the lookup ID. It will always try to locate the K closest nodes in the network and starts by sending
	// udp messages and recursively locates nodes that are closer until no more nodes can be found. Each node
	// will then receive these messages and search through their own routing table

	// Get the initial k closest nodes from the current node
	kClosestNodes := network.localNode.routingTable.FindClosestContacts(&lookupID, network.k)

	// Keep sending RPCs until k closest nodes has been visited
	for network.VisitedKClosest(kClosestNodes) {
		// Grab <=alpha nodes to visit
		var nodesToVisit []Contact
		alphaCount := 0
		for i := 0; i < network.k; i++ {
			if kClosestNodes[i].visited == false { // If node hasn't been visited yet, lets do so
				kClosestNodes[i].visited = true
				nodesToVisit = append(nodesToVisit, kClosestNodes[i])
				alphaCount += 1
				if alphaCount == network.alpha { // We got alpha nodes to visit now, so lets not grab anymore
					break
				}
			}
		}

		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for i := 0; i < network.alpha; i++ {
			// TODO: Do this asynchronously?
			newBucket := network.SendFindContactMessage(&nodesToVisit[i], lookupID) // Send RPC
			// Update the k closest nodes list
			// We do this by appending new nodes (non-duplicates) to the list
			// and then sorting the list based on distance to the lookupID parameter
			kClosestNodes = network.UpdateKClosest(kClosestNodes, newBucket)
		}
	}

	return kClosestNodes
}

func (network *Network) UpdateKClosest(kClosestNodes []Contact, newNodes []Contact) []Contact {
	for i := 0; i < network.k; i++ { // For each node in newNodes
		curNode := newNodes[i]
		alreadyIn := false
		// Check if the new node is already part of k closest
		for j := 0; j < network.k; j++ {
			if kClosestNodes[j].ID == curNode.ID {
				alreadyIn = true
				break
			}
		}
		// If it isnt, add to kClosest. Otherwise do nothing
		if !alreadyIn {
			kClosestNodes = append(kClosestNodes, curNode)
		}
	}

	return network.SortNodes(kClosestNodes)
}

// Insertion sorting a list of k closest nodes based on their contact.distance
func (network *Network) SortNodes(kClosestNodes []Contact) []Contact {
	for i := 1; i < len(kClosestNodes); i++ {
		curNode := kClosestNodes[i]
		for j := i; j >= 0; j-- {
			if j == 0 || curNode.distance.Less(kClosestNodes[j].distance) {
				kClosestNodes[i] = kClosestNodes[j]
				kClosestNodes[j] = curNode
			}
		}
	}
	return kClosestNodes
}

func (network *Network) VisitedKClosest(kClosestNodes []Contact) bool {
	for i := 0; i < network.k; i++ {
		if kClosestNodes[i].visited == false {
			return false
		}
	}
	return true
}

func (network *Network) SendFindContactMessage(contact *Contact, targetID KademliaID) []Contact {
	conn, err := net.Dial("udp", contact.Address + ":5001")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
	} else {
		fmt.Println("Connection established! Sending ping msg.")
	}

	// setup payload properly
	payload := make([]byte, 1)
	payload[0] = FIND_NODE
	payload = append(payload, targetID[:]...)
	conn.Write(payload)

	reply := make([]byte, 1 + 20 + BUCKET_DATA_SIZE) // TODO: Whats the size of 1 bucket in bytes? Its constant at least :)
	conn.Read(reply)
	var kClosestReply bucket
	json.Unmarshal(reply[21:], kClosestReply)
	return kClosestReply.GetContactsAndCalcDistances(&targetID)
}

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
			// TODO Send a STORE RPC and make it wörk
			go func() {
				conn, err := net.Dial("udp", contact.Address + ":5001")
				fmt.Println("Connection established!")

				if err != nil {
					fmt.Println("Could not establish connection to " + contact.ID.String())
				} else {
					conn.Write(storeMessage)
				}
				conn.Close()
			}()
		}
	}
}
