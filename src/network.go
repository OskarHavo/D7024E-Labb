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

const MAX_PACKET_SIZE = 1024
const IP_LEN = 4
const HEADER_LEN = 1
const BUCKET_HEADER_LEN = 1
const TIMEOUT = 500

const KAD_PORT = "5001"

type Network struct {
	localNode Node
}

func NewNetwork(ip *net.IP) Network {
	return Network{NewNode(NewContact(NewKademliaIDFromIP(ip),ip.String()))}
}

// Handles FIND_NODE  requests (initiated by findNodeRPC) from other nodes by sending back a bucket of the k closest
// nodes to some kademlia ID.
// msgType is the type of message (message description) that will be sent back to the requester.
func (network *Network) sendFindNodeAck(msg *[]byte, connection *net.UDPConn, address *net.UDPAddr, msgType byte) {
	// Message format:
	// REC: [MSG TYPE, REQUESTER ID, TARGET ID]
	// SEND: [MSG TYPE, BUCKET SIZE, BUCKET:[ID, IP]]

	requesterID := (*KademliaID)((*msg)[HEADER_LEN:HEADER_LEN+ID_LEN])
	targetID := (*KademliaID)((*msg)[HEADER_LEN+ID_LEN : HEADER_LEN+ID_LEN+ID_LEN])
	bucket := network.localNode.routingTable.FindClosestContacts(targetID, k + 1)
	bucket = removeSelfOrTail(requesterID, bucket, len(bucket) == k + 1)

	fmt.Println("Received a FIND_NODE request from node", requesterID,
		"with a target ID", targetID)

	// 1 byte for type of msg, 1 byte for number of contacts
	var reply = make([]byte, HEADER_LEN+BUCKET_HEADER_LEN+(ID_LEN+IP_LEN)*len(bucket))

	// Set the message type
	reply[0] = msgType

	// Set the length of the bucket to send back
	reply[HEADER_LEN] = byte(len(bucket))

	// Send the actual bucket (serialize the contacts and put them in the message)
	for i ,data := range bucket {
		// static size is the size that is the same for all replies
		// (size of msg type + size of bucket + size of my ID)
		staticSize := HEADER_LEN + BUCKET_HEADER_LEN
		// dynamic size is the size from prev loops (size of i-1 serialized contacts in bucket)
		dynamicSize := (ID_LEN + IP_LEN) * i

		// Set node ID
		copy(reply[staticSize+dynamicSize : staticSize+dynamicSize+(ID_LEN+IP_LEN) ], data.ID[:])

		// convert ip string to byte array
		var nodeAddress = net.ParseIP(strings.Split(data.Address,":")[0]).To4()

		// Put the IP address
		copy(reply[staticSize+dynamicSize+ID_LEN: staticSize+dynamicSize+ID_LEN+IP_LEN],
			nodeAddress)
	}
	(*connection).WriteToUDP(reply, address)
}

// unpackMessage handles all kademlia requests from other nodes.
func (network *Network) unpackMessage(msg *[]byte, connection *net.UDPConn, address *net.UDPAddr) {
	switch messageType := (*msg)[0]; messageType {
	case PING:
		requesterID := (*KademliaID)((*msg)[HEADER_LEN:HEADER_LEN+ID_LEN])

		fmt.Println("Received a PING request from node", requesterID.String())
		reply := make([]byte, HEADER_LEN)
		reply[0] = PING_ACK

		_,err := (*connection).WriteToUDP(reply, address)
		if err != nil {
			fmt.Println("There was an error when replying to a PING request.", err.Error())
		}
		return
	case FIND_NODE:
		network.sendFindNodeAck(msg, connection, address, FIND_NODE_ACK)
		return
	case FIND_DATA:
		// Message format:
		// REC:  [MSG TYPE, REQUESTER ID, HASH]
		// SEND: [MSG TYPE, REQUESTER ID, BUCKET SIZE, BUCKET:[ID, IP]]
		//   OR  [MSG TYPE, DATA]
		fmt.Println("Received a FIND_DATA request")
		hash := (*KademliaID)((*msg)[HEADER_LEN+ID_LEN : HEADER_LEN+ID_LEN+ID_LEN])
		data := network.localNode.LookupData(hash)
		if data != nil {
			var reply = make([]byte, HEADER_LEN+len(data))
			reply[0] = FIND_DATA_ACK_SUCCESS
			copy(reply[HEADER_LEN:], data)
			_, err := (*connection).WriteToUDP(reply, address)
			if err != nil {
				fmt.Println("There was an error when replying to a FIND_DATA request.", err.Error())
			}
		} else {
			network.sendFindNodeAck(msg, connection, address, FIND_DATA_ACK_FAIL)
		}
		return
	case STORE:
		// Message format:
		// REC: [MSG TYPE, REQUESTER ID, HASH, DATA...]
		// SEND: nothing
		requesterID := (*KademliaID)((*msg)[HEADER_LEN:HEADER_LEN+ID_LEN])
		hash := (*KademliaID)((*msg)[HEADER_LEN+ID_LEN:HEADER_LEN+ID_LEN+ID_LEN])
		data := (*msg)[HEADER_LEN+ID_LEN+ID_LEN:MAX_PACKET_SIZE]
		fmt.Println("Received a STORE request from node", requesterID.String())

		network.localNode.Store(data, hash)
		return
	}
}

// Listen listens for incoming requests. Once a message is received it is directed to unpackMessage.
// Also checks if the requesting node should be added to the routing table of the local node
// (see kickTheBucket)
func (network *Network) Listen() {
		for {
			conn, err := net.ListenUDP("udp", &net.UDPAddr{
				Port:5001,
			})
			if err == nil {
				msg := make([]byte, MAX_PACKET_SIZE)
				conn.SetReadDeadline(time.Now().Add(TIMEOUT * time.Millisecond))
				_,addr,err := conn.ReadFromUDP(msg)

				if err != nil {
					fmt.Println("Could not read incoming message")
				} else {
					ID := (*KademliaID)(msg[1:1+ID_LEN])


					//fmt.Println("Attempting to create a contact")
					contact := NewContact(ID,addr.IP.To4().String())
					//fmt.Println("Attempting to kick the bucket")
					network.kickTheBucket(&contact)

					network.unpackMessage(&msg,conn,addr)
				}

			} else {
				fmt.Println("Could not read from incoming connection.", err.Error())
			}
			conn.Close()
		}
}

// Join a kademlia network via a known nodes IP and ID. The ID is probably the SHA-1 hash of its IP.
func (network *Network) Join(id *KademliaID, address string) {
	knownNode := NewContact(id, address)
	//network.localNode.routingTable.AddContact(knownNode)

	if network.Ping(&knownNode) { // If Ping is successful
		fmt.Println("Joined network node " + knownNode.Address + " successfully!")
		network.NodeLookup(network.localNode.routingTable.me.ID) // Start lookup algorithm on yourself
	}
}

// Ping some node directly with the given contact.address.
// Returns true if the node responded successfully, and false if it did not
func (network *Network) Ping(contact *Contact) bool {
	hostName := contact.Address
	service := hostName + ":" + KAD_PORT

	start := time.Now()
	remoteAddr, err := net.ResolveUDPAddr("udp",service)
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		return false
	} else {
		fmt.Println("Connection established to " + remoteAddr.String() + "!")
	}

	// Setup msg and send
	msg := make([]byte, HEADER_LEN+ID_LEN)
	msg[0] = PING
	copy(msg[HEADER_LEN:], network.localNode.routingTable.me.ID[:])
	conn.Write(msg)

	// Setup and read reply
	msg = make([]byte, HEADER_LEN)
	conn.SetReadDeadline(time.Now().Add(20*time.Second))
	conn.ReadFromUDP(msg)

	tmp := make([]byte,255)
	//conn.Read(tmp)
	conn.SetReadDeadline(time.Now().Add(TIMEOUT * time.Millisecond))
	_,_,err2 := conn.ReadFromUDP(tmp)
	conn.Close()

	if err2 != nil {
		fmt.Println("Could not read Ping message from", contact.ID.String())
		return false
	}

	duration := time.Since(start)

	// Update routing table with the contact that we pinged
	network.kickTheBucket(contact)

	if msg[0] == PING_ACK {
		fmt.Println("Successful ping to " + contact.ID.String() + " took " + strconv.FormatInt(duration.Milliseconds(),
			10) + " ms")
		return true
	} else {
		fmt.Println("Received unrecognized response from node", contact.ID.String(), "when pinged")
		fmt.Println("Received message of type " + strconv.FormatInt(int64(msg[0]),10))
		return false
	}
}

// NodeLookup is the central kademlia node lookup algorithm that can be used to find nodes (or data, see DataLookup)
// depending on the lookup ID. It will always try to locate the K closest nodes in the network and starts by sending
// udp messages and recursively locates nodes that are closer until no more nodes can be found. Each node
// will then receive these messages and search through their own routing table
func (network *Network) NodeLookup(lookupID *KademliaID) []Contact {
	// Get the initial k closest nodes from the current node
	initNodes := network.localNode.routingTable.FindClosestContacts(lookupID, k)
	if len(initNodes) == 0 {
		return []Contact{}
	}

	var visited ContactCandidates
	var unvisited ContactCandidates
	unvisited.Append(initNodes)

	wideSearch := false
	var searchRange = alpha
	for !visitedKClosest(&unvisited, &visited, k) { // Keep sending RPCs until k closest nodes has been visited
		searchRange = setSearchSize(wideSearch, &unvisited)

		var newRoundNodes []Contact
		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for currentNode := 0; currentNode < searchRange; {
			newBucket, success := network.findNodeRPC(&unvisited.contacts[currentNode], lookupID) // Send RPC
			if success {
				newRoundNodes = append(newRoundNodes, newBucket...)
				currentNode ++
			} else {
				unvisited.contacts = append(unvisited.contacts[:currentNode],
					unvisited.contacts[currentNode+1:]...)
				searchRange--
			}
		}
		postIterationProcessing(&visited, &unvisited, &newRoundNodes, searchRange)
	}

	if visited.Len() < k {
		return visited.GetContacts(visited.Len())
	} else {
		return visited.GetContacts(k)
	}
}

// DataLookup works exactly like NodeLookup, except that we return data instead of a bucket if we find it from
// any of the findDataRPCs (which replaces findNodeRPC from NodeLookup)
func (network *Network) DataLookup(hash *KademliaID) ([]byte, []Contact) {
	localData := network.localNode.LookupData(hash)
	if localData != nil {
		fmt.Println("Found data on local node")
		return localData, []Contact{network.localNode.routingTable.me}
	}

	initNodes := network.localNode.routingTable.FindClosestContacts(hash, k)
	if len(initNodes) == 0 {
		return nil, []Contact{}
	}

	var visited ContactCandidates
	var unvisited ContactCandidates
	unvisited.Append(initNodes)

	wideSearch := false
	var searchRange = alpha
	for !visitedKClosest(&unvisited, &visited, k) {
		searchRange = setSearchSize(wideSearch, &unvisited)

		var newRoundNodes []Contact
		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for currentNode := 0; currentNode < searchRange; {
			data, newBucket, success := network.findDataRPC(&unvisited.contacts[currentNode], hash) // Send RPC
			if success {
				if data != nil {
					return data, unvisited.contacts[currentNode:currentNode+1]
				}
				newRoundNodes = append(newRoundNodes, newBucket...)
				currentNode ++
			} else {
				unvisited.contacts = append(unvisited.contacts[:currentNode],
					unvisited.contacts[currentNode+1:]...)
				searchRange--
			}
		}
		postIterationProcessing(&visited, &unvisited, &newRoundNodes, searchRange)
	}

	if visited.Len() < k {
		return nil, visited.GetContacts(visited.Len())
	} else {
		return nil, visited.GetContacts(k)
	}
}

// Store sends a store msg to the 20th closest nodes a bucket
func (network *Network) Store(data []byte, hash *KademliaID) {
	var nodes = network.NodeLookup(hash) // Get ALL nodes that are closest to the hash value
	fmt.Println("Storing data in " + strconv.FormatInt(int64(len(nodes)),10) + " external nodes")
	network.localNode.routingTable.me.CalcDistance(hash)
	if len(nodes) < k {
		nodes = append(nodes, network.localNode.routingTable.me)
	} else if network.localNode.routingTable.me.distance.Less(nodes[len(nodes)-1].distance) {
		// If the locals node distance is less than the last node in the bucket,
		// Im actually supposed to be in the bucket and not that node.
		nodes[len(nodes)-1] = network.localNode.routingTable.me
	}
	for _,contact := range nodes { // What type of syntax is this??
		if network.localNode.routingTable.me.ID == contact.ID {
			// No need to send a network request. Send the RPC directly to the local node thread.
			fmt.Println("Storing data on local node")
			network.localNode.Store(data, hash)
		} else {
			// This is easily done async because we don't have to care what happens after!
			go network.storeDataRPC(contact, hash, data)
		}
	}
}

// findNodeRPC sends a FIND_NODE request to some contact with some targetID.
// Returns the k closest nodes to the target ID and if the connection to the contact was successful or not
func (network *Network) findNodeRPC(contact *Contact, targetID *KademliaID) ([]Contact, bool) {
	hostName := contact.Address
	service := hostName + ":" + KAD_PORT
	remoteAddr, err := net.ResolveUDPAddr("udp",service)

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Could not establish connection when sending findNodeRPC to ", contact.ID.String(),"   ", contact.Address)
		return nil,false
	} else {
		// Message format:
		// SEND: [MSG TYPE, REQUESTER ID, TARGET ID]
		// REC:  [MSG TYPE, REQUESTER ID, BUCKET SIZE, BUCKET:[ID, IP]]

		// Send FIND_NODE request
		msg := make([]byte, HEADER_LEN+ID_LEN+ID_LEN)
		msg[0] = FIND_NODE
		copy(msg[HEADER_LEN : HEADER_LEN+ID_LEN], network.localNode.routingTable.me.ID[:])
		copy(msg[HEADER_LEN+ID_LEN: HEADER_LEN+ID_LEN+ID_LEN], targetID[:])
		conn.Write(msg)

		// Read and handle reply
		reply := make([]byte, HEADER_LEN+BUCKET_HEADER_LEN+(ID_LEN+IP_LEN)*k)
		conn.SetReadDeadline(time.Now().Add(TIMEOUT * time.Millisecond))
		_,_,err := conn.ReadFromUDP(reply)

		conn.Close()

		if err != nil {
			fmt.Println("Could not read FIND_NODE_RPC from " + contact.ID.String())
			return nil,false
		}

		// TODO: This can be put into a function and reused in findDataRPC
		kClosestReply := handleBucketReply(&reply)

		network.kickTheBucket(contact)
		return kClosestReply.GetContactsAndCalcDistances(targetID), true
	}
}

// findNodeRPC sends a FIND_DATA request to some contact with some targetID.
// Returns the k closest nodes to the hash OR the data that matches the hash (in the hash, data pair)
// and if the connection to the contact was successful or not. If the connection was unsuccessful,
// both data and k closest contacts are nil.
func (network *Network) findDataRPC(contact *Contact, hash *KademliaID) ([]byte, []Contact, bool) {
	hostName := contact.Address
	service := hostName + ":" + KAD_PORT
	remoteAddr, err := net.ResolveUDPAddr("udp",service)

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Could not establish connection when sending findDataRPC to " + contact.ID.String())
		return nil, nil, false
	} else {
		fmt.Println("Sending FIND_DATA to node ", contact.ID.String())

		msg := make([]byte, HEADER_LEN+ID_LEN+ID_LEN)
		msg[0] = FIND_DATA
		copy(msg[HEADER_LEN : HEADER_LEN+ID_LEN], network.localNode.routingTable.me.ID[:])
		copy(msg[HEADER_LEN+ID_LEN: HEADER_LEN+ID_LEN+ID_LEN], hash[:])
		conn.Write(msg)

		reply := make([]byte, MAX_PACKET_SIZE)
		conn.SetReadDeadline(time.Now().Add(TIMEOUT * time.Millisecond))
		_,_,err := conn.ReadFromUDP(reply)

		conn.Close()

		if err != nil {
			fmt.Println("Could not read FIND_DATA_RPC from " + contact.ID.String())
			return nil, nil, false
		}

		// TODO This updates the routing table with the node we just queried.
		network.kickTheBucket(contact)

		if reply[0] == FIND_DATA_ACK_FAIL {
			// Message format:
			// REC: [MSG TYPE, REQUESTER ID, BUCKET SIZE, BUCKET:[ID, IP]]
			// (This has the same format as findNodeAck)
			kClosestReply := handleBucketReply(&reply)
			return nil, kClosestReply.GetContactsAndCalcDistances(hash), true

		} else if reply[0] == FIND_DATA_ACK_SUCCESS {
			// Message format:
			// REC: [MSG TYPE, DATA]
			return reply[HEADER_LEN:], nil, true
		} else {
			return nil, nil, false
		}
	}
}

// storeDataRPC sends a STORE request to some contact with a hash value and some data
// The function does not care if the data is correctly stored or not by the contact
// and therefore does not return anything
func (network *Network) storeDataRPC(contact Contact, hash *KademliaID, data []byte) {
	hostName := contact.Address
	service := hostName + ":" + KAD_PORT
	remoteAddr, err := net.ResolveUDPAddr("udp",service)
	conn, err := net.DialUDP("udp", nil, remoteAddr)

	if err != nil {
		fmt.Println("Could not establish connection when sending storeDataRPC to " + contact.ID.String())
	} else {
		// Message format:
		// SEND: [MSG TYPE, REQUESTER ID, HASH, DATA...]
		// REC: nothing

		// Prepare STORE RPC
		storeMessage := make([]byte, HEADER_LEN+ID_LEN)
		storeMessage[0] = STORE
		copy(storeMessage[HEADER_LEN:HEADER_LEN+ID_LEN], network.localNode.routingTable.me.ID[:])
		copy(storeMessage[HEADER_LEN+ID_LEN:HEADER_LEN+ID_LEN+ID_LEN], hash[:])
		storeMessage = append(storeMessage, data...)

		conn.Write(storeMessage)
	}
	conn.Close()
}

// Check if a bucket is full and then kick one node if it does not respond to a ping message.
// Call this function whenever you want to add a new node to the routing table. The node can either
// already exist in a bucket or be a new node.
func (network *Network) kickTheBucket(contact *Contact) {
	// First find the appropriate bucket
	bucketIndex := network.localNode.routingTable.getBucketIndex(contact.ID)
	bucket := network.localNode.routingTable.buckets[bucketIndex]

	if bucket.Len() == k {
		element := bucket.Contains(contact)
		if element != nil {
			bucket.list.MoveToFront(element)
		} else {
			// Choose a node to sacrifice
			sacrifice := bucket.list.Back().Value.(Contact)

			if network.Ping(&sacrifice) {
				//fmt.Println("Received ping from sacrifice node. Node was not kicked from the bucket.")
			} else {
				bucket.list.Remove(bucket.list.Back())
				bucket.AddContact(*contact)
			}
		}
	} else {
		bucket.AddContact(*contact)
	}
}

// We don't want to send back the requester its own ID so that it has itself in its own bucket.
// removeSelfOrTail therefore grabs a bucket (of size k+1) and either remove the requesterID if it exists,
// or the tail (the furthest one away of the nodes) if it doesn't.
// Removing tail is optional (you don't want to do this if bucket is already less than k)
func removeSelfOrTail(requesterID *KademliaID, bucket []Contact, removeTail bool) []Contact {
	for index, contact := range bucket {
		if *requesterID == *contact.ID {
			bucket = append(bucket[:index], bucket[index + 1:]...)
			return bucket
		} else {
		}
	}
	if removeTail {
		return bucket[:len(bucket)-1]
	}
	return bucket
}

// addNewNodes adds new nodes from the current iteration to the unvisited collection
// It avoids duplicates, which means that all nodes in unvisited + visited will be unique (ID vise)
func addNewNodes(visited *ContactCandidates, unvisited *ContactCandidates,
	newNodes []Contact) {
	allOld := *visited // All nodes from the previous rounds that we have seen, visited and unvisited
	allOld.Append(unvisited.contacts)
	var toBeAdded ContactCandidates
	for i := 0; i < len(newNodes); i++ {
		// Check for duplicates among the nodes from prev rounds (visited and unvisited)
		// Check for duplicates among newNodes
		if !allOld.Contains(&newNodes[i]) && !toBeAdded.Contains(&newNodes[i]) {
			toBeAdded.AppendContact(newNodes[i])
			//toBeAdded.Sort()
		}
	}
	unvisited.Append(toBeAdded.contacts)
	//unvisited.Sort()
}

// visitedKClosest checks if the NodeLookup (and DataLookup) algorithm is finished by
// comparing the known k closest nodes to the visited nodes
// returns either true (finished) or false (not finished). See implementation comments for more detail
func visitedKClosest(unvisited *ContactCandidates, visited *ContactCandidates, k int) bool {
	visited.Sort()
	unvisited.Sort()

	// There are no new contacts to visit! We must be done, regardless of how many nodes
	// we have already visited (name of function should change?)
	if unvisited.Len() == 0 {
		return true
	}

	// If we have visited k or more nodes, we could POTENTIALLY be done. Let's dive deeper ...
	if visited.Len() >= k {
		// If the last contact in visited is closer than the first contact in unvisited,
		// all visited contacts are the closest. We are done.
		if visited.contacts[k-1].Less(&unvisited.contacts[0]) {
			return true
		}
		// Otherwise, there is a contact that is part of the k closest collection that
		// we have not yet visited.
		return false
	} else {
		// If we have nodes to visit (unvisited.Len() > 0) but we have not yet visited k nodes,
		// we are not done
		return false
	}
}

// doWideSearch checks if the next iteration in the NodeLookup (or DataLookup) algorithm should be a
// "wide search" or not by comparing the distance of some new contacts (this iteration)
// to some already known contacts (previous iterations)
// Definition of wide search: "If a round of FIND_NODEs fails to return a node any closer than the closest already seen,
// the initiator resends the FIND_NODE to all of the closest k nodes it has not already queried"
func doWideSearch(newContacts *[]Contact, closest Contact) bool {
	for _, contact := range *newContacts {
		if contact.Less(&closest) {
			return false
		}
	}
	return true
}

// handleBucketReply takes a byte slice and unserializes it into a bucket (collection of contacts)
func handleBucketReply(msg *[]byte) bucket {
	totalContacts := int((*msg)[HEADER_LEN])
	result := *newBucket()
	for i := 0; i < totalContacts; i++ {
		// static size is the size that is the same for all replies
		// (size of msg type + size of bucket + size of my ID)
		staticSize := HEADER_LEN + BUCKET_HEADER_LEN
		// dynamic size is the size from prev loops (size of i-1 serialized contacts in bucket)
		dynamicSize := (ID_LEN + IP_LEN) * i

		id := (*msg)[staticSize+dynamicSize: staticSize+dynamicSize+ID_LEN]

		IP := net.IPv4((*msg)[staticSize+dynamicSize+ID_LEN],
			(*msg)[staticSize+dynamicSize+ID_LEN+1],
			(*msg)[staticSize+dynamicSize+ID_LEN+2],
			(*msg)[staticSize+dynamicSize+ID_LEN+3])
		contact := NewContact((*KademliaID)(id), IP.String())
		result.AddContact(contact)
	}
	return result
}

// setSearchSize returns the number of nodes to visit this iteration.
// The size is dependent on the boolean wideSearch (if wide search is enabled or not)
// and how many unvisited nodes there are
func setSearchSize(wideSearch bool, unvisitedNodes *ContactCandidates) int {
	var result int
	if wideSearch {
		// Grab <=k nodes to visit
		wideSearch = false
		result = k
	} else {
		// Grab <=alpha nodes to visit
		result = alpha
	}
	if result > unvisitedNodes.Len() {
		result = unvisitedNodes.Len()
	}
	return result
}

// postIterationProcessing does all the things that is required after an iteration of contacting
// nodes is completed in NodeLookup (or DataLookup). This includes:
// 		moving visited nodes from the unvisited collection to the visited collection
// 		calling addNewNodes
//      calling and returning the result of doWideSearch
func postIterationProcessing(visited *ContactCandidates, unvisited *ContactCandidates,
	newRoundNodes *[]Contact, searchRange int) bool {
	if unvisited.Len() > 0 {
		visited.Append(unvisited.contacts[:searchRange])
		visited.Sort()

		wideSearch := doWideSearch(newRoundNodes, visited.contacts[0])

		unvisited.contacts = unvisited.contacts[searchRange:]
		addNewNodes(visited, unvisited, *newRoundNodes)
		return wideSearch
	}
	return false
}