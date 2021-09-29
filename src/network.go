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
	localNode Node
}

func NewNetwork(ip *net.IP) Network {
	return Network{NewNode(NewContact(NewKademliaIDFromIP(ip),ip.String()))}
}
// TODO - dokumentation.
func (network *Network) sendFindNodeAck(msg *[]byte, connection *net.UDPConn, address *net.UDPAddr, msgType byte) {
	requesterID := (*KademliaID)((*msg)[1:1+IDLength])
	targetID := (*KademliaID)((*msg)[1+IDLength:1+IDLength+IDLength])
	bucket := network.localNode.routingTable.FindClosestContacts(targetID, k + 1)
	bucket = removeSelfOrTail(requesterID, bucket, len(bucket) == k + 1)

	// Prepare reply
	// [TYPE MSG, REQUESTER ID, BUCKET SIZE, BUCKET:[ID, IP]]
	IPLength := 4
	msgTypeLength := 1
	sizeBucketLength := 1

	// 1 byte msg, 20 byte ID, 1 byte for number of contacts
	var reply = make([]byte, msgTypeLength+IDLength+sizeBucketLength+(IDLength+IPLength)*len(bucket))

	// Set the message type
	reply[0] = msgType

	// Set the requester ID (my ID)
	copy(reply[msgTypeLength:msgTypeLength+IDLength], network.localNode.routingTable.me.ID[:])

	// Set the length of the bucket to send back
	reply[msgTypeLength+IDLength] = byte(len(bucket))

	// Send the actual bucket (serialize the contacts and put them in the message)
	for i ,data := range bucket {
		// static size is the size that is the same for all replies
		// (size of msg type + size of bucket + size of my ID)
		staticSize := msgTypeLength + IDLength + sizeBucketLength
		// dynamic size is the size from prev loops (size of i-1 serialized contacts in bucket)
		dynamicSize := (IDLength + IPLength) * i
		//copy(reply[2+IDLength+(IDLength+4)*i : 2+(IDLength+4)*i+IDLength],data.ID[:])
		// Set node ID
		copy(reply[staticSize+dynamicSize : staticSize+dynamicSize+(IDLength+IPLength) ], data.ID[:])

		// convert ip string to byte array
		var nodeAddress = net.ParseIP(strings.Split(data.Address,":")[0]).To4()
		fmt.Println("IP Address of node in findNodeAck: " + nodeAddress.String())

		//copy(reply[(2+IDLength+IDLength)+(IDLength+4)*i : (2+IDLength+IDLength)+(IDLength+4)*i+4],
			//nodeAddress[12:])
		// Put the IP address
		fmt.Println("Im giving this IP back to the requester:", nodeAddress.String())
		copy(reply[staticSize+dynamicSize+IDLength : staticSize+dynamicSize+IDLength+IPLength],
			nodeAddress)

		fmt.Println("Computed the following IP address to send:\n")
		for i := 0; i < 4; i++ {
			fmt.Print(strconv.FormatInt(int64(reply[staticSize+dynamicSize+IDLength+i]),10))
			fmt.Print(".")
		}
		fmt.Print("\n")
		for i := 0; i < 4; i++ {
			fmt.Print(strconv.FormatInt(int64(nodeAddress[i]),10))
			fmt.Print(".")
		}
	}
	(*connection).WriteToUDP(reply, address)
}

// Server receive function for network messages
func (network *Network) unpackMessage(msg *[]byte, connection *net.UDPConn, address *net.UDPAddr) {
	switch message_type := (*msg)[0]; message_type {
	case PING:
		fmt.Println("Received a ping message. Sending ack")
		reply := make([]byte, 1+IDLength)
		reply[0] = PING_ACK

		copy(reply[1:],network.localNode.routingTable.me.ID[:])

		// TODO DOES NOT WORK??
		//(*connection).Write(reply)
		_,err := (*connection).WriteToUDP(reply,address)
		if err != nil {
			fmt.Println("There was a ping error: " + err.Error())
		}
		return
	case PING_ACK:
		//fmt.Println("Received PING ACK!!!!!!")
		return
	case STORE:
		fmt.Println("Received a store message.")
		network.localNode.Store((*msg)[1+IDLength:], (*KademliaID)((*msg)[1:1+IDLength]))
		return
	case STORE_ACK:
		// TODO I dunno, send a message to the GUI or something.
		return
	case FIND_NODE:
		fmt.Println("Received a find node message. Sending ack")
		network.sendFindNodeAck(msg,connection,address, FIND_NODE_ACK)
		return
	case FIND_DATA:
		fmt.Println("Received a find data message. Sending ack")
		ID := (*KademliaID)((*msg)[1:1+IDLength])
		result := network.localNode.LookupData(ID)
		if result != nil {
			var reply = make([]byte,1+IDLength+len(result))
			reply[0] = FIND_DATA_ACK_SUCCESS
			copy(reply[1:IDLength+1],network.localNode.routingTable.me.ID[:])
			copy(reply[IDLength+1:],result)
			(*connection).Write(reply)
		} else {
			network.sendFindNodeAck(msg,connection,address,FIND_DATA_ACK_FAIL)
		}
		return
	}
}

// Listen for incoming connections
func (network *Network) Listen() {
	//hostName := "localhost"
	//portNum := "5001"
	//service := hostName + ":" + portNum
	//addr,err := net.ResolveUDPAddr("udp4",service)

	//if err != nil {
	//	fmt.Println(err)
	//} else {
		for {
			fmt.Println("Listening to UDP")
			conn, err := net.ListenUDP("udp", &net.UDPAddr{
				Port:5001,
			})
			if err == nil {
				msg := make([]byte, 1024)
				_,addr,_ := conn.ReadFromUDP(msg)
				fmt.Println("Received message: ")
				ID := (*KademliaID)(msg[1:1+IDLength])
				fmt.Println("    ID: " + ID.String())
				fmt.Println("    IP: " + addr.String())


				fmt.Println("Attempting to create a contact")
				contact := NewContact(ID,addr.String())
				fmt.Println("Attempting to kick the bucket")
				network.kickTheBucket(&contact)

				network.unpackMessage(&msg,conn,addr)
			} else {
				fmt.Println("Could not read from incoming connection")
			}
			conn.Close()
		}

	//}

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
// Returns if true if the node responded successfully, and false if it did not
func (network *Network) Ping(contact *Contact) bool {
	hostName := contact.Address
	portNum := "5001"

	service := hostName + ":" + portNum

	remoteAddr, err := net.ResolveUDPAddr("udp",service)

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		return false
	} else {
		fmt.Println("Connection established to " + remoteAddr.String() + "!")
	}

	reply := make([]byte, 1+IDLength)
	reply[0] = PING
	copy(reply[1:],network.localNode.routingTable.me.ID[:])

	start := time.Now()
	//conn.Write(reply)
	conn.Write(reply)

	tmp := make([]byte,255)
	conn.SetReadDeadline(time.Now().Add(20*time.Second)) // TODO Change to something more appropriate
	//conn.Read(tmp)
	conn.ReadFromUDP(tmp)
	conn.Close()

	duration := time.Since(start)

	network.kickTheBucket(contact)

	if tmp[0] == PING_ACK {
		fmt.Println("Successful ping to " + contact.ID.String() + " took " + strconv.FormatInt(duration.Milliseconds(),
			10) + " ms")
		return true
	} else {
		fmt.Println("Received unrecognized response from node " + (*KademliaID)(tmp[1:]).String() + " when pinged\n" +
			"    Received message of type " + strconv.FormatInt(int64(tmp[0]),10))
		return false
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

	if len(initNodes) == 0 {
		return []Contact{}
	}

	fmt.Println("Starting find Node with " + strconv.FormatInt(int64(len(initNodes)),10) + " nodes")
	unvisited.Append(initNodes)

	wideSearch := false
	var dynamic_alpha = alpha
	for !visitedKClosest(unvisited, visited, k) { // Keep sending RPCs until k closest nodes has been visited
		//var nodesToVisit []Contact
		if wideSearch {
			// Grab <=k nodes to visit
			wideSearch = false
			dynamic_alpha = k
		} else {
			// Grab <=alpha nodes to visit
			dynamic_alpha = alpha
		}
		if dynamic_alpha > unvisited.Len() {
			dynamic_alpha = unvisited.Len()
		}

		var newRoundNodes []Contact
		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		fmt.Println("\n\nStarting a new iteration.")


		for k := 0; k < dynamic_alpha; {
			newBucket := network.findNodeRPC(&unvisited.contacts[k], lookupID) // Send RPC
			if newBucket != nil {
				newRoundNodes = append(newRoundNodes, newBucket...)
			} else {
				unvisited.contacts = append(unvisited.contacts[:k], unvisited.contacts[k+1:]...)
				k ++
			}
		}

		// "If a round of FIND_NODEs fails to return a node any closer than the closest already seen, the initiator
		// resends the FIND_NODE to all of the closest k nodes it has not already queried" <-- we call this a
		// "wide search
		wideSearch = doWideSearch(newRoundNodes, unvisited.contacts[0])

		// Move unvisited nodes to the visited.
		visited.Append(unvisited.contacts[:dynamic_alpha])
		visited.Sort()
		unvisited.contacts = unvisited.contacts[dynamic_alpha:]

		// Prepare more unvisited nodes.
		network.updateKClosest(&visited, &unvisited, newRoundNodes)
	}
	fmt.Println("finished node lookup with " + strconv.FormatInt(int64(visited.Len()),10) + " nodes")
	if visited.Len() < k {
		return visited.GetContacts(visited.Len())
	} else {
		return visited.GetContacts(k)
	}
}

// DataLookup works exactly like NodeLookup, except that we return data instead of a bucket if we find it from
// any of the findDataRPCs (which replaces findNodeRPC from NodeLookup)
// TODO- Gör så den här fungerar på samma sätt som NodeLookup. Jag har tagit bort nodesToVisit. Den skapade mer problem än nytta.
func (network *Network) DataLookup(hash *KademliaID) ([]byte, []Contact) {

	local_data := network.localNode.LookupData(hash)
	if local_data != nil {
		return local_data, []Contact{network.localNode.routingTable.me}
	}

	var visited ContactCandidates
	var unvisited ContactCandidates
	initNodes := network.localNode.routingTable.FindClosestContacts(hash, k)
	unvisited.Append(initNodes)


	var data []byte
	var newBucket []Contact
	wideSearch := false
	for !visitedKClosest(unvisited, visited, k) {
		var nodesToVisit []Contact
		if wideSearch {
			// Grab <=k nodes to visit
			nodesToVisit = unvisited.GetContacts(k)
			wideSearch = false
		} else {
			// Grab <=alpha nodes to visit
			nodesToVisit = unvisited.GetContacts(alpha)
		}

		var newRoundNodes []Contact
		// Actually visit <=alpha of k-closest nodes grabbed in the prev step
		for i := 0; i < len(nodesToVisit); i++ {
			data, newBucket = network.findDataRPC(&nodesToVisit[i], hash, &visited, &unvisited) // Send RPC
			newRoundNodes = append(newRoundNodes, newBucket...)
			if data != nil {
				return data, []Contact{nodesToVisit[i]}
			}
		}
		network.updateKClosest(&visited, &unvisited, newRoundNodes)

		// "If a round of FIND_NODEs fails to return a node any closer than the closest already seen, the initiator
		// resends the FIND_NODE to all of the closest k nodes it has not already queried" <-- we call this a
		// "wide search
		wideSearch = doWideSearch(newRoundNodes, nodesToVisit[0])
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
	fmt.Println("Storing data in " + strconv.FormatInt(int64(len(nodes)),10) + " nodes")
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
// TODO- Dokumentation
func (network *Network) findNodeRPC(contact *Contact, targetID *KademliaID) []Contact {
	hostName := contact.Address
	portNum := "5001" // TODO STATIC CONST
	service := hostName + ":" + portNum
	remoteAddr, err := net.ResolveUDPAddr("udp",service)

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Could not establish connection when sending findNodeRPC to " + contact.ID.String())
		return nil
	} else {

		// Prepare msg
		// [TYPE MSG, REQUESTER ID, BUCKET SIZE, BUCKET:[ID, IP]]
		IPLength := 4
		msgTypeLength := 1
		sizeBucketLength := 1

		msg := make([]byte, msgTypeLength+IDLength+IDLength)
		msg[0] = FIND_NODE
		copy(msg[msgTypeLength : msgTypeLength+IDLength], network.localNode.routingTable.me.ID[:])
		copy(msg[msgTypeLength+IDLength : msgTypeLength+IDLength+IDLength], contact.ID[:])
		conn.Write(msg)

		reply := make([]byte, msgTypeLength+IDLength+sizeBucketLength+(IDLength+IPLength)*k)
		conn.ReadFromUDP(reply)
		conn.Close()

		totalContacts := int(reply[msgTypeLength+IDLength])
		kClosestReply := newBucket()
		for i := 0; i < totalContacts;i++ {
			// static size is the size that is the same for all replies
			// (size of msg type + size of bucket + size of my ID)
			staticSize := msgTypeLength + IDLength + sizeBucketLength
			// dynamic size is the size from prev loops (size of i-1 serialized contacts in bucket)
			dynamicSize := (IDLength + IPLength) * i

			id := reply[staticSize+dynamicSize: staticSize+dynamicSize+IDLength]

			IP := net.IPv4(reply[staticSize+dynamicSize+IDLength],
				reply[staticSize+dynamicSize+IDLength+1],
				reply[staticSize+dynamicSize+IDLength+2],
				reply[staticSize+dynamicSize+IDLength+3])
			contact := NewContact((*KademliaID)(id), IP.String())

			kClosestReply.AddContact(contact)
		}

		// TODO This updates the routing table with the node we just queried.
		network.kickTheBucket(contact)
		return kClosestReply.GetContactsAndCalcDistances(targetID)
	}
}
// TODO- Dokumentation
func (network *Network) findDataRPC(contact *Contact, hash *KademliaID,
	visited *ContactCandidates, unvisited *ContactCandidates) ([]byte, []Contact) {
	hostName := contact.Address
	portNum := "5001"
	service := hostName + ":" + portNum
	remoteAddr, err := net.ResolveUDPAddr("udp",service)
	conn, err := net.DialUDP("udp", nil, remoteAddr)

	if err != nil {
		fmt.Println("Could not establish connection when sending findDataRPC to " + contact.ID.String())

		return nil, nil
	} else {
		fmt.Println("Connection established to " + contact.ID.String() + "!")
		fmt.Println("Sending findDataRPC ...")

		// We are visiting the node, so we move it from unvisited to visited collection
		unvisited.Remove(contact)
		visited.AppendContact(*contact)

		payload := make([]byte, 1+IDLength)
		payload[0] = FIND_DATA
		copy(payload[1:1+IDLength],network.localNode.routingTable.me.ID[:])
		payload = append(payload, hash[:]...)
		conn.Write(payload)

		reply := make([]byte,1+IDLength+1+(IDLength+4)*k)
		conn.ReadFromUDP(reply)
		conn.Close()

		// TODO This updates the routing table with the node we just queried.
		network.kickTheBucket(contact)

		if reply[0] == FIND_DATA_ACK_FAIL {
			totalContacts := int(reply[1+IDLength])
			var kClosestReply bucket
			for i := 0; i < totalContacts;i++ {
				id := reply[2+IDLength+(IDLength+4)*i:2+(IDLength+4)*i+IDLength]
				IP := net.IP{}
				copy(IP[12:],reply[2+IDLength+(IDLength+4)*i+IDLength:2+(IDLength+4)*i+IDLength+4])
				contact := NewContact((*KademliaID)(id),IP.String())

				kClosestReply.AddContact(contact)
			}
			return nil, kClosestReply.GetContactsAndCalcDistances(hash)
		} else if reply[0] == FIND_DATA_ACK_SUCCESS {
			// FIND_DATA_ACK_SUCCESS
			// Success!!
			fmt.Println("Received data from node " + contact.ID.String() + "!")
			return reply[1:],nil
		} else {
			fmt.Println("Received unrecognized response from node " + contact.ID.String() + " during findDataRPC")
			return nil, nil
		}
	}
}

// TODO- Dokumentation
func (network *Network) storeDataRPC(contact Contact, hash *KademliaID, data []byte) {
	hostName := contact.Address
	portNum := "5001"
	service := hostName + ":" + portNum
	remoteAddr, err := net.ResolveUDPAddr("udp",service)
	conn, err := net.DialUDP("udp", nil, remoteAddr)

	if err != nil {
		fmt.Println("Could not establish connection when sending storeDataRPC to " + contact.ID.String())
	} else {
		fmt.Println("Connection established to " + contact.ID.String() + "!")
		fmt.Println("Sending storeDataRPC ...")

		// Prepare STORE RPC
		storeMessage := make([]byte, 1+IDLength)
		storeMessage[0] = STORE
		copy(storeMessage[1:],network.localNode.routingTable.me.ID[:])
		storeMessage = append(storeMessage, hash[:]...)
		storeMessage = append(storeMessage, data...)

		conn.Write(storeMessage)
	}
	conn.Close()
}

// We don't want to send back the requester its own ID so that it has itself in its own bucket.
// removeSelfOrTail therefore grabs a bucket of size k+1 and either remove the requesterID if it exists,
// or the tail (the furthest one away of the 21 nodes) if it doesn't.
// Removing tail is optional (you don't want to do this if bucket is already less than k)
func removeSelfOrTail(requesterID *KademliaID, bucket []Contact, removeTail bool) []Contact {
	for index, contact := range bucket {
		//fmt.Println("[RSOT] RequesterID:", requesterID.String())
		//fmt.Println("[RSOT] contactID:", contact.ID.String())
		if *requesterID == *contact.ID {
			//fmt.Println("[RSOT] These are the same!")
			bucket = append(bucket[:index], bucket[index + 1:]...)
			return bucket
		} else {
			//fmt.Println("[RSOT] These are not the same!")
		}
	}
	if removeTail {
		return bucket[:len(bucket)-1]
	}
	return bucket
}

// updateKClosest updates the list of the k closest nodes used in the NodeLookup algorithm
// It is done by appending new nodes (non-duplicates) to the unvisited collection
func (network *Network) updateKClosest(visited *ContactCandidates, unvisited *ContactCandidates,
	newNodes []Contact) {
	allOld := *visited // All nodes from the previous rounds that we have seen, visited and unvisited
	allOld.Append(unvisited.contacts)
	fmt.Println("Updating K closest. Old nodes: ", allOld.Len())
	var toBeAdded ContactCandidates
	for i := 0; i < len(newNodes); i++ {
		// Check for duplicates among the nodes from prev rounds (visited and unvisited)
		// Check for duplicates among newNodes
		if !allOld.Contains(&newNodes[i]) && !toBeAdded.Contains(&newNodes[i]){
			toBeAdded.AppendContact(newNodes[i])
			toBeAdded.Sort()
		}
	}
	unvisited.Append(toBeAdded.contacts)
	unvisited.Sort()
}
// TODO- Dokumentation
func visitedKClosest(unvisited ContactCandidates, visited ContactCandidates, k int) bool {
	visited.Sort()
	unvisited.Sort()

	// There are no new contacts to visit! We must be done, regardless of how many nodes
	// we have already visited
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

func doWideSearch(newContacts []Contact, closest Contact) bool {
	for _, contact := range newContacts {
		if contact.Less(&closest) {
			return false
		}
	}
	return true
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
				fmt.Println("Received ping from sacrifice node. Node was not kicked from the bucket.")
			} else {
				fmt.Println("Updating node " + contact.ID.String() + " to routing table!")
				bucket.list.Remove(bucket.list.Back())
				bucket.AddContact(*contact)
			}
		}
	} else {
		bucket.AddContact(*contact)
	}
}