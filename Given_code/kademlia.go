package d7024e

type Node struct {
	storage map[KademliaID][]byte
}

func newNode() Node {
	return Node{make(map[KademliaID][]byte)}
}

func (kademlia *Node) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Node) LookupData(hash KademliaID) {
	// TODO
}

func (kademlia *Node) Store(data []byte, hash KademliaID) {

	kademlia.storage[hash] = data
}
