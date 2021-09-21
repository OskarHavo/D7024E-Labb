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
	FIND_DATA_ACK_SUCCESS byte = iota
	FIND_DATA_ACK_FAIL byte = iota
)

type Network struct {
	localNode *Node
}

func NewNetwork(node *Node) Network {
	return Network{node}
}

func (network *Network) sendFindNodeAck(msg *[]byte, connection *net.Conn, msgType byte) {
	requesterID := (*KademliaID)((*msg)[1:1+IDLength])
	targetID := (*KademliaID)((*msg)[1+IDLength:1+IDLength+IDLength])
	bucket := network.localNode.routingTable.FindClosestContacts(targetID, k + 1)
	bucket = removeSelfOrTail(requesterID, bucket)


	// Prepare reply
	var reply = make([]byte,1+1+(IDLength+4)*len(bucket)) // 1 byte msg, 1 byte for number of contacts
	reply[0] = msgType // Set the message type
	reply[1] = byte(len(bucket))

	// Serialize the contacts and put them in the message
	var i = 0
	for _,data := range bucket {
		copy(reply[2+(IDLength+4)*i:2+(IDLength+4)*i+IDLength],data.ID[:])
		var address = net.ParseIP(data.Address)[12:]
		copy(reply[(2+IDLength)+(IDLength+4)*i:(2+IDLength)+(IDLength+4)*i+4],address)
		i++
	}

	// Final structure of message:
	// FIND_NODE_ACK + number of contacts + (ID + IP) + (ID + IP) + ...

	(*connection).Write(reply)
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
		network.localNode.Store((*msg)[1+IDLength:], (*KademliaID)((*msg)[1:1+IDLength]))
		return
	case STORE_ACK:
		// TODO I dunno, send a message to the GUI or something.
		return
	case FIND_NODE:
		/*
		requesterID := (*KademliaID)((*msg)[1:1+IDLength])
		targetID := (*KademliaID)((*msg)[1+IDLength:1+IDLength+IDLength])
		bucket := network.localNode.routingTable.FindClosestContacts(targetID, k + 1)
		bucket = removeSelfOrTail(requesterID, bucket)


		// Prepare reply
		var reply = make([]byte,1+1+(IDLength+4)*len(bucket)) // 1 byte msg, 1 byte for number of contacts
		reply[0] = FIND_NODE_ACK // Set the message type
		reply[1] = byte(len(bucket))

		// Serialize the contacts and put them in the message
		var i = 0
		for _,data := range bucket {
			copy(reply[2+(IDLength+4)*i:2+(IDLength+4)*i+IDLength],data.ID[:])
			var address = net.ParseIP(data.Address)[12:]
			copy(reply[(2+IDLength)+(IDLength+4)*i:(2+IDLength)+(IDLength+4)*i+4],address)
			i++
		}

		// Final structure of message:
		// FIND_NODE_ACK + number of contacts + (ID + IP) + (ID + IP) + ...

		(*connection).Write(reply)
		 */
		network.sendFindNodeAck(msg,connection, FIND_NODE_ACK)
		return
	case FIND_DATA:
		ID := (*KademliaID)((*msg)[1:1+IDLength])
		result := network.localNode.LookupData(ID)
		if result != nil {
			var reply = make([]byte,1+len(result))
			reply[0] = FIND_DATA_ACK_SUCCESS
			copy(reply[1:],result)
			(*connection).Write(reply)
		} else {
			network.sendFindNodeAck(msg,connection,FIND_DATA_ACK_FAIL)
		}
		return
	}
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


func (network *Network) NodeLookup(lookupID *KademliaID) []Contact {
	// This will be the central node lookup algorithm that can be used to find nodes or data depending on
	// the lookup ID. It will always try to locate the K closest nodes in the network and starts by sending
	// udp messages and recursively locates nodes that are closer until no more nodes can be found. Each node
	// will then receive these messages and search through their own routing table

	// Get the initial k closest nodes from the current node

	var visited ContactCandidates
	var unvisited ContactCandidates
	initNodes := network.localNode.routingTable.FindClosestContacts(lookupID, k)
	unvisited.Append(initNodes)

	wideSearch := false
	for visitedKClosest(unvisited, visited, k) { // Keep sending RPCs until k closest nodes has been visited
		var nodesToVisit []Contact
		if wideSearch {
			// Grab <=k nodes to visit
			nodesToVisit = unvisited.GetContacts(k)
			wideSearch = false
		} else {
			// Grab <=alpha nodes to visit
			nodesToVisit = unvisited.GetContacts(alpha)
		}

		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for i := 0; i < len(nodesToVisit); i++ {
			// TODO: Do this asynchronously?
			newBucket := network.findNodeRPC(&nodesToVisit[i], lookupID, &visited, &unvisited) // Send RPC

			//oldKClosest := unvisited
			network.updateKClosest(&visited, &unvisited, newBucket)
			//newKClosest := unvisited
			// "If a round of FIND_NODEs fails to return a node any closer than the closest already seen, the initiator
			// resends the FIND_NODE to all of the closest k nodes it has not already queried" <-- we call this a
			// "wide search"
			/* TODO
			for _, contact := range newBucket {
				wideSearch = doWideSearch()
			}
			if oldKClosest.Equals(newKClosest, k) {
				wideSearch = doWideSearch()
			}
			*/
		}
	}

	return visited.GetContacts(k)
}


func doWideSearch(newContacts []Contact, visited ContactCandidates, unvisited ContactCandidates) {
	visited.Append(unvisited.contacts)
	for _, contact := range newContacts {
		//if contact TODO
	}
}

// DataLookup works exactly like NodeLookup, except that we return data instead of a bucket if we find it from
// any of the findDataRPCs (which replaces findNodeRPC from NodeLookup)
func (network *Network) DataLookup(hash *KademliaID) ([]byte, []Contact) {
	var visited ContactCandidates
	var unvisited ContactCandidates
	initNodes := network.localNode.routingTable.FindClosestContacts(hash, k)
	unvisited.Append(initNodes)


	var data []byte
	var newBucket []Contact
	wideSearch := false
	for visitedKClosest(unvisited, visited, k) {
		var nodesToVisit []Contact
		if wideSearch {
			// Grab <=k nodes to visit
			nodesToVisit = unvisited.GetContacts(k)
			wideSearch = false
		} else {
			// Grab <=alpha nodes to visit
			nodesToVisit = unvisited.GetContacts(alpha)
		}

		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for i := 0; i < len(nodesToVisit); i++ {
			data, newBucket = network.findDataRPC(&nodesToVisit[i], hash) // Send RPC
			if data != nil {
				return data, nil
			}

			//oldKClosest := closestNodes
			network.updateKClosest(&visited, &unvisited, newBucket)
			//newKClosest := closestNodes
			// "If a round of FIND_NODEs fails to return a node any closer than the closest already seen, the initiator
			// resends the FIND_NODE to all of the closest k nodes it has not already queried" <-- we call this a
			// "wide search"
			/* TODO
			if oldKClosest.Equals(newKClosest, k) {
				wideSearch = true
			} */
		}
	}
	return nil, visited.GetContacts(k)
}

// Store sends a store msg to the 20th closest nodes a bucket
func (network *Network) Store(data []byte, hash *KademliaID) {
	var nodes = network.NodeLookup(hash) // Get ALL nodes that are closest to the hash value
	network.localNode.routingTable.me.CalcDistance(hash)
	if network.localNode.routingTable.me.distance.Less(nodes[len(nodes)-1].distance) {
		// If the locals node distance is less than the last node in the bucket,
		// Im actually supposed to be in the bucket and not that node.
		nodes[len(nodes)-1] = network.localNode.routingTable.me
	}
	for _,contact := range nodes { // What type of syntax is this??
		if network.localNode.routingTable.me.ID == contact.ID {
			// No need to send a network request. Send the RPC directly to the local node thread.
			network.localNode.Store(data, hash)
		} else {
			// This is easily done async because we don't have to care what happens after!
			go network.storeDataRPC(contact, hash, data)
		}
	}
}


func (network *Network) findNodeRPC(contact *Contact, targetID *KademliaID,
	visited *ContactCandidates, unvisited *ContactCandidates) []Contact {

	conn, err := net.Dial("udp", contact.Address + ":5001")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())

		// "Nodes that fail to respond (quickly) are removed from
		// consideration untilk and unless they do respond"
		unvisited.Remove(contact)

		return nil
	} else {
		fmt.Println("Connection established! Sending find node RPC.")

		// We are visiting the node, so we move it from unvisited to visited collection
		unvisited.Remove(contact)
		visited.AppendContact(*contact)

		reply := make([]byte,1+1+(IDLength+4)*k)
		conn.Read(reply)
		totalContacts := int(reply[1])
		var kClosestReply bucket
		for i := 0; i < totalContacts;i++ {
			id := reply[2+(IDLength+4)*i:2+(IDLength+4)*i+IDLength]
			IP := net.IP{}
			copy(IP[12:],reply[2+(IDLength+4)*i+IDLength:2+(IDLength+4)*i+IDLength+4])
			contact := NewContact((*KademliaID)(id),IP.String())

			kClosestReply.AddContact(contact)
		}
		return kClosestReply.GetContactsAndCalcDistances(targetID)
	}
}

func (network *Network) findDataRPC(contact *Contact, hash *KademliaID) ([]byte, []Contact) {
	conn, err := net.Dial("udp", contact.Address + ":5001")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
	} else {
		fmt.Println("Connection established! Sending ping msg.")
	}

	payload := make([]byte, 1)
	payload[0] = FIND_DATA
	payload = append(payload, network.localNode.routingTable.me.ID[:]...)
	payload = append(payload, hash[:]...)
	conn.Write(payload)

	reply := make([]byte,1+1+(IDLength+4)*k)
	conn.Read(reply)

	if reply[0] == FIND_DATA_ACK_FAIL {
		totalContacts := int(reply[1])
		var kClosestReply bucket
		for i := 0; i < totalContacts;i++ {
			id := reply[2+(IDLength+4)*i:2+(IDLength+4)*i+IDLength]
			IP := net.IP{}
			copy(IP[12:],reply[2+(IDLength+4)*i+IDLength:2+(IDLength+4)*i+IDLength+4])
			contact := NewContact((*KademliaID)(id),IP.String())

			kClosestReply.AddContact(contact)
		}
		return nil, kClosestReply.GetContactsAndCalcDistances(hash)
	} else {
		// FIND_DATA_ACK_SUCCESS
		// Success!!
		return reply[1:],nil
	}

}


func (network *Network) storeDataRPC(contact Contact, hash *KademliaID, data []byte) {
	conn, err := net.Dial("udp", contact.Address + ":5001")
	fmt.Println("Connection established!")

	if err != nil {
		fmt.Println("Could not establish connection to " + contact.ID.String())
	} else {
		// Prepare STORE RPC
		storeMessage := make([]byte, 1)
		storeMessage[0] = STORE
		storeMessage = append(storeMessage, hash[:]...)
		storeMessage = append(storeMessage, data...)

		conn.Write(storeMessage)
	}
	conn.Close()
}

// We don't want to send back the requester its own ID so that it has itself in its own bucket.
// removeSelfOrTail therefore grabs a bucket of size k+1 and either remove the requesterID if it exists,
// or the tail (the furthest one away of the 21 nodes) if it doesn't.
func removeSelfOrTail(requesterID *KademliaID, bucket []Contact) []Contact {
	for index, contact := range bucket {
		if requesterID == contact.ID {
			bucket = append(bucket[:index], bucket[index + 1:]...)
			return bucket
		}
	}
	return bucket[:len(bucket)-1]
}

// updateKClosest updates the list of the k closest nodes used in the NodeLookup algorithm
// It is done by appending new nodes (non-duplicates) to the unvisited collection
// and then sorting the list based on distance to the lookupID parameter
func (network *Network) updateKClosest(visited *ContactCandidates, unvisited *ContactCandidates,
	newNodes []Contact) {
	all := *visited
	all.Append(unvisited.contacts)
	var toBeAdded []Contact
	for i := 0; i < len(newNodes); i++ {
		if !all.Contains(&newNodes[i]) {
			toBeAdded = append(toBeAdded, newNodes[i])
		}
	}
	unvisited.Append(toBeAdded)
	unvisited.Sort()
}

func visitedKClosest(unvisited ContactCandidates, visited ContactCandidates, k int) bool {
	if visited.Len() < k { // Cant have visited k closest if we haven't even visited k nodes yet
		return false
	}

	visited.Append(unvisited.contacts)
	visited.Sort()
	unvisited.Sort()

	for i := 0; i < k; i++ {
		if visited.contacts[i].ID != unvisited.contacts[i].ID {
			return false
		}
	}
	return true
}