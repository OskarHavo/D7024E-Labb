package main

import (
	"fmt"
	"sync"
)

// The node itself is an object that runs on it's own thread and waits for commands from the networking part
// of a container. We don't need to perform any udp calls from here, just return messages to the local
// network thread which then sends it through the network. Okidoki?
type Node struct {
	storage map[KademliaID][]byte
	routingTable *RoutingTable

	// map that contains the time-to-live of each data object in storage
	ttl map[KademliaID]int

	// a list of maximum k contacts that should be refreshed for some data object in storage until
	// it is forgotten
	refreshContacts map[KademliaID][]Contact
	refreshMutex sync.Mutex

	storateMutex sync.Mutex
}

// Create a new Node
func NewNode(ID Contact) Node {
	return Node{make(map[KademliaID][]byte), NewRoutingTable(ID),
		make(map[KademliaID]int), make(map[KademliaID][]Contact), sync.Mutex{},sync.Mutex{}}
}

// Local lookup of the size closest contacts to some target kademlia ID
func (kademlia *Node) LookupContact(target *KademliaID, size int) []Contact {
	return kademlia.routingTable.FindClosestContacts(target, size)
}

// Lookup data
func (kademlia *Node) LookupData(hash *KademliaID) []byte {
	kademlia.storateMutex.Lock()
	defer kademlia.storateMutex.Unlock()
	if kademlia.storage[*hash] == nil {
		return nil
	}
	return kademlia.storage[*hash]
}

// Store data
func (kademlia *Node) Store(data []byte, hash *KademliaID) {
	kademlia.storateMutex.Lock()
	defer kademlia.storateMutex.Unlock()
	if  kademlia.storage[*hash] != nil{
		// TODO Throw some error or something
		return
	}
	kademlia.storage[*hash] = data
	kademlia.ttl[*hash] = TIME_TO_LIVE
}

// Delete data stored at some hash
func (kademlia *Node) Delete(hash *KademliaID) {
	kademlia.storateMutex.Lock()
	defer kademlia.storateMutex.Unlock()
	if kademlia.storage[*hash] != nil { // Only delete if there is actually something there
		delete(kademlia.storage, *hash) // Delete the data
		delete(kademlia.ttl, *hash) // Delete the ttl associated with the data
	}
}

// Refresh will update the ttl associated with some data by setting it to some system-wide
// predetermined parameter
func (kademlia *Node) Refresh(hash *KademliaID) {
	if kademlia.ttl[*hash] != 0 { // Can't refresh something that is already dead
		kademlia.ttl[*hash] = TIME_TO_LIVE
	} else {
		fmt.Println("ERROR! Trying to locally refresh something that is already dead. Hash is:",
			hash.String())
	}
}

// Forget will remove the contacts associated to some data hash, which means no more refreshRPCs
// will be sent to those contacts and the data will eventually be deleted by the contacts, including
// "this node" if it is one of the associated ones
func (kademlia *Node) Forget(hash *KademliaID) {
	kademlia.refreshMutex.Lock()
	defer kademlia.refreshMutex.Unlock()
	if len(kademlia.refreshContacts[*hash]) != 0 {
		delete(kademlia.refreshContacts, *hash)
	}
}

// RememberContacts remembers which contacts are associated to some data hash
// so that they can be refreshed in the future
func (kademlia *Node) RememberContacts(hash *KademliaID, contacts []Contact) {
	kademlia.refreshMutex.Lock()
	defer kademlia.refreshMutex.Unlock()
	if len(kademlia.refreshContacts[*hash]) == 0 {
		kademlia.refreshContacts[*hash] = contacts
	}
}
