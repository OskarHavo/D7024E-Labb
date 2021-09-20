package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Network requests:
// 3 bit (1 byte) message description:
// 000: PING
// 010: STORE
// 100: FIND_NODE
// 110: FIND_VALUE
// Followed by a 20 byte ID of needed and data for STORE command
// 2 bit + 20 byte +

// Protocol for returning information:
// PING_ACK: Contains nothing

// STORE_ACK: Not sure if this is a requirement. We can probably skip it

// FIND_NODE_ACK: Nodes are stored in tuples with <IP, NODE_ID> in a long list without description of how many nodes there are.
// 		We already know the size of each tuple and the size of the array -> size/tuple_bytes = number of tuples

// FIND_VALUE_ACK: message type followed by one byte indicating a list of nodes or some actual data.
// 		0: Found no data. Returns <=K closest nodes
// 		1: Found data. Returns the full byte array

// Golang doesn't have enums, this the closest alternative I could find
const (
	PING byte = iota
	PING_ACK byte = iota

	STORE byte = iota
	STORE_ACK byte = iota

	FIND_NODE byte = iota
	FIND_NODE_ACK byte = iota

	FIND_DATA byte = iota
	FIND_DATA_ACK byte = iota
)

type Network struct {
	localNode *Node
}

func NewNetwork(node *Node) Network {
	return Network{node}
}

// Server receive function for network messages
func (network *Network) unpackMessage(msg *[]byte, connection *net.Conn) {
	switch id := (*msg)[0]; id {
	case PING:
		if (*msg)[1] == 0{
			var msg = []byte{PING,1} // PING + ACK
			(*connection).Write(msg)
		} else {
			// TODO notify the user interface of the ping message.
		}
		return
	case STORE:
		msg
		return
	case FIND_NODE:
		// TODO: Needs to ignore sending back the recipient to itself!!
		requesterID := (*msg)[1:20]
		targetID := (*msg)[21:31] // HELP ROBIN! How 2 extract ID from byte array???
		bucket := network.localNode.routingTable.FindClosestContacts(targetID, k + 1)
		bucket = removeSelfOrTail(requesterID, bucket)
		var reply = []byte{FIND_NODE, 1}
		reply = append(reply, bucket[:]...) // HELP ROBIN! Send bucket back!!!
		(*connection).Write()
		return
	case FIND_DATA:
		return
	}
}

// We dont wanna send back the requester its own ID so that it has itself in its own bucket.
// Therefore we grab a bucket of k+1 and either remove the requesterID if it exists, or the tail (the furthest one away
// of the 21 nodes) if it doesn't.
func removeSelfOrTail(requesterID KademliaID, bucket []Contact) []Contact {
	for index, contact := range bucket {
		if requesterID == *contact.ID {
			bucket = append(bucket[:index], bucket[index + 1:]...)
			return bucket
		}
	}
	return bucket[:len(bucket)-1]
}


// Listen for incoming connections
func (network *Network) Listen(ip string, port string) {
	conn, err := net.Listen("udp", port)
	if err != nil {
		fmt.Println(err)
	} else {
		for {
			c, err := conn.Accept()
			if err == nil {
				var msg []byte = nil
				c.Read(msg)
			}
		}
	}
}

func (network *Network) Ping(contact *Contact) {
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

func (network *Network) FindNode(lookupID KademliaID) []Contact {
	// This will be the central node lookup algorithm that can be used to find nodes or data depending on
	// the lookup ID. It will always try to locate the K closest nodes in the network and starts by sending
	// udp messages and recursively locates nodes that are closer until no more nodes can be found. Each node
	// will then receive these messages and search through their own routing table

	// Get the initial k closest nodes from the current node
	var closestNodes ContactCandidates
	nodes := network.localNode.routingTable.FindClosestContacts(&lookupID, k)
	closestNodes.Append(nodes)

	// Keep sending RPCs until k closest nodes has been visited
	for closestNodes.Visited(k) {
		// Grab <=alpha nodes to visit
		nodesToVisit := closestNodes.GetUnvisited(alpha)

		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for i := 0; i < len(nodesToVisit); i++ {
			// TODO: Do this asynchronously?
			/* TODO: newBucket has to have each distance calculated towards lookupID!!!
					 (to be done on the response side (or here ...)
			         newBucket CANNOT have duplicates (duh ...) */
			newBucket := network.sendLookupMessage(&nodesToVisit[i], lookupID) // Send RPC
			// Update the k closest nodes list
			// We do this by appending new nodes (non-duplicates) to the list
			// and then sorting the list based on distance to the lookupID parameter
			network.updateKClosest(&closestNodes, newBucket)
		}
	}

	return closestNodes.GetContacts(k)
}

func (network *Network) updateKClosest(kClosestNodes *ContactCandidates, newNodes []Contact) {
	var toBeAdded []Contact
	for i := 0; i < len(newNodes); i++ {
		if !kClosestNodes.Contains(&newNodes[i]) {
			toBeAdded = append(toBeAdded, newNodes[i])
		}
	}
	kClosestNodes.Append(toBeAdded)
	kClosestNodes.Sort()
}

func (network *Network) sendLookupMessage(contact *Contact, targetID KademliaID) []Contact {
	conn, err := net.Dial("udp", contact.Address + ":5001")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
	} else {
		fmt.Println("Connection established! Sending ping msg.")
	}

	// setup payload properly - HELP ROBIN!
	payload := make([]byte, 1)
	payload[0] = FIND_NODE
	payload = append(payload, network.localNode.routingTable.me.ID[:]...)
	payload = append(payload, targetID[:]...)
	conn.Write(payload)

	reply := make([]byte, 1 + 20) // TODO: Whats the size of 1 bucket in bytes? Its constant at least :)
	conn.Read(reply)
	var kClosestReply bucket
	return kClosestReply.GetContactsAndCalcDistances(&targetID)
}

// DataLookup Works exactly like NodeLookup, except that we return data if we receive any from
func (network *Network) FindData(hash KademliaID) ([]byte, []Contact) {
	var closestNodes ContactCandidates
	nodes := network.localNode.routingTable.FindClosestContacts(&hash, k)
	closestNodes.Append(nodes)

	var data []byte
	var newBucket []Contact
	for closestNodes.Visited(k) {
		// Grab <=alpha nodes to visit
		nodesToVisit := closestNodes.GetUnvisited(alpha)

		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for i := 0; i < len(nodesToVisit); i++ {
			data, newBucket = network.sendDataMessage(&nodesToVisit[i], hash) // Send RPC
			if data != nil {
				return data, nil
			}
			network.updateKClosest(&closestNodes, newBucket)
		}
	}
	return nil, closestNodes.GetContacts(k)
}

func (network *Network) sendDataMessage(contact *Contact, hash KademliaID) ([]byte, []Contact) {
	conn, err := net.Dial("udp", contact.Address + ":5001")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
	} else {
		fmt.Println("Connection established! Sending ping msg.")
	}

	// setup payload properly - HELP ROBIN!
	payload := make([]byte, 1)
	payload[0] = FIND_DATA
	payload = append(payload, network.localNode.routingTable.me.ID[:]...)
	payload = append(payload, hash[:]...)
	conn.Write(payload)

	reply := make([]byte, 1 + 20) // TODO: Reply is gonna be EITHER a bucket or the data we are looking for!
	// HELP ROBIN!
	conn.Read(reply)
	//var kClosestReply bucket
	return nil, nil
	//return nil, bucket
	//return data, nil
}

// Man ska ju inte ge en hash, utan bara data och sen ger kademlia tillbaka en hash? Eller görs det lokalt i Node
// structen i kademlia.go innan?
// SendStoreMessage sends a store msg to the 20th closest nodes a bucket
func (network *Network) Store(hash KademliaID, data []byte) {
	// Prepare STORE RPC
	storeMessage := make([]byte, 1)
	storeMessage[0] = STORE
	storeMessage = append(storeMessage, hash[:]...)
	storeMessage = append(storeMessage, data...)

	var nodes = network.FindNode(hash) // Get ALL nodes that are closest to the hash value
	for _,contact := range nodes { // What type of syntax is this??
		if network.localNode.routingTable.me.ID == contact.ID {
			// No need to send a network request. Send the RPC directly to the local node thread.
		} else {
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
