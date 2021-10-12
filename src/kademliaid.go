package main

import (
	"crypto/sha1"
	"encoding/hex"
	"net"
)

// the static number of bytes in a KademliaID
const ID_LEN = 20

// type definition of a KademliaID
type KademliaID [ID_LEN]byte

func sha1Hash(content string) []byte {
	h := sha1.New()
	h.Write([]byte(content))
	hashedFileBytes := h.Sum(nil)
	return hashedFileBytes
}

// NewKademliaID returns a new instance of a KademliaID based on the string input
func NewKademliaID(data string) *KademliaID {
	decoded, _ := hex.DecodeString(data)

	newKademliaID := KademliaID{}
	for i := 0; i < ID_LEN; i++ {
		newKademliaID[i] = decoded[i]
	}

	return &newKademliaID
}
// NewKademliaIDFromData returns a new instance of a KademliaID based on the hash input
func NewKademliaIDFromData(data string) *KademliaID {
	decoded := sha1Hash(data)

	newKademliaID := KademliaID{}
	for i := 0; i < ID_LEN; i++ {
		newKademliaID[i] = decoded[i]
	}

	return &newKademliaID
}

// Create kademlia ID from an IP address, for example a node.
func NewKademliaIDFromIP(ip *net.IP) *KademliaID {
	decoded := sha1Hash(ip.String())
	newKademliaID := KademliaID{}
	for i := 0; i < ID_LEN; i++ {
		newKademliaID[i] = decoded[i]
	}

	return &newKademliaID
}

// Less returns true if kademliaID < otherKademliaID (bitwise)
func (kademliaID KademliaID) Less(otherKademliaID *KademliaID) bool {
	for i := 0; i < ID_LEN; i++ {
		if kademliaID[i] != otherKademliaID[i] {
			return kademliaID[i] < otherKademliaID[i]
		}
	}
	return false
}

// Equals returns true if kademliaID == otherKademliaID (bitwise)
func (kademliaID KademliaID) Equals(otherKademliaID *KademliaID) bool {
	for i := 0; i < ID_LEN; i++ {
		if kademliaID[i] != otherKademliaID[i] {
			return false
		}
	}
	return true
}

// CalcDistance returns a new instance of a KademliaID that is built 
// through a bitwise XOR operation betweeen kademliaID and target
func (kademliaID KademliaID) CalcDistance(target *KademliaID) *KademliaID {
	result := KademliaID{}
	for i := 0; i < ID_LEN; i++ {
		result[i] = kademliaID[i] ^ target[i]
	}
	return &result
}

// String returns a simple string representation of a KademliaID
func (kademliaID *KademliaID) String() string {
	return hex.EncodeToString(kademliaID[0:ID_LEN])
}
