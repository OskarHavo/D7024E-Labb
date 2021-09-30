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

// William skulle fixa
func (kademlia *Node) LookupContact(target *Contact) {
	// TODO - Vad ska den här göra? Behövs den?
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
