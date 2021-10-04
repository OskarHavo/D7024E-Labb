package main

// The node itself is an object that runs on it's own thread and waits for commands from the networking part
// of a container. We don't need to perform any udp calls from here, just return messages to the local
// network thread which then sends it through the network. Okidoki?
type Node struct {
	storage map[KademliaID][]byte
	routingTable *RoutingTable
}

// Create a new Node
func NewNode(ID Contact) Node {
	return Node{make(map[KademliaID][]byte), NewRoutingTable(ID)}
}

// Local lookup of the size closest contacts to some target kademlia ID
func (kademlia *Node) LookupContact(target *KademliaID, size int) []Contact {
	return kademlia.routingTable.FindClosestContacts(target, size)
}

// Lookup data
func (kademlia *Node) LookupData(hash *KademliaID) []byte {
	if kademlia.storage[*hash] == nil {
		return nil
	}
	return kademlia.storage[*hash]
}

// Store data
func (kademlia *Node) Store(data []byte, hash *KademliaID) {
	if  kademlia.storage[*hash] != nil{
		// TODO Throw some error or something
		return
	}
	kademlia.storage[*hash] = data
}
