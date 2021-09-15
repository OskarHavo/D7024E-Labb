package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const alpha = 3

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

func (network *Network) nodeLookupRPC(ID KademliaID) []Contact {
	// this function does the actual network stuff. It should return the closest nodes in the bucket
	return nil
}

func (network *Network) nodeLookup(ID KademliaID) []Contact {

	// TODO
	// This will be the central node lookup algorithm that can be used to find nodes or data depending on
	// the lookup ID. It will always try to locate the K closest nodes in the network and starts by sending
	// udp messages and recursively locates nodes that are closer until no more nodes can be found. Each node
	// will then receive these messages and search through their own routing table

	// Step 1: call FIND_NODE on the local node. Get K of the closest nodes we know of so far
	// Step 2: While we have not queried all possible nodes:
	// 		If there are new nodes that are closer:
	//			Call the alpha closest nodes that we have not queried yet.
	// 		Else:
	// 			Call the K closest nodes that we have not queried yet.

	// Requirements:
	// - Keep track of queried nodes by storing all nodes in a sorted list and removing nodes that have been queried.
	// - Keep track of the K closest nodes in a list and update it as needed. Also set an update flag.

	findMessage := make([]byte,1)
	findMessage[0] = FIND_NODE
	findMessage = append(findMessage, ID[:]...)

	// Send a findNode message to the local node
	network.localNode.comChannel <- findMessage

	// This collects local nodes to start the search with
	// All nodes will be sorted and have a precalculated distance to the target

	var contactQueue = ContactCandidates{<-network.localNode.contactChannel}

	// We need to also put these notes in a list of candidates
	var closestContacts = contactQueue
	var variableAlpha = alpha
	for ; contactQueue.Len() > 0; {
		var currentQueue []Contact
		if variableAlpha >= len(contactQueue.contacts) {
			currentQueue = contactQueue.contacts
			contactQueue.contacts = contactQueue.contacts
		} else {
			currentQueue = contactQueue.contacts[:variableAlpha]
			contactQueue.contacts = contactQueue.contacts[variableAlpha:]
		}

		var rpcResult = []Contact{}
		for _,cnt := range currentQueue {
			// TODO Send FIND_NODE RPC to cnt and receive the result in rpcResult
			fmt.Println(cnt.ID.String()) // To disable compiler error. Delete this whenever possible
		}
		// TODO Wait for the results or something before continuing
		var flag = false
		for _, newContact := range rpcResult {
			if !closestContacts.Contains(&newContact) {
				closestContacts.Append([]Contact{newContact})
				contactQueue.contacts = append(contactQueue.contacts, newContact)
				flag = true
			}
		}
		if flag {
			// Wee we found more nodes. Let's sort the lists again for another round
			closestContacts.Sort()
			contactQueue.Sort()
			variableAlpha = alpha
		} else {
			variableAlpha = bucketSize
		}
	}

	if closestContacts.Len() <= bucketSize {
		return closestContacts.contacts
	} else {
		return closestContacts.contacts[:bucketSize]
	}
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

	var nodes = network.nodeLookup(hash) // Get ALL nodes that are closest to the hash value
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

	var nodes = network.nodeLookup(hash) // Get ALL nodes that are closest to the hash value
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
