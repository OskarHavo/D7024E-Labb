package main

import (
	"fmt"
	"time"
)

const (
	TIME_TO_LIVE = 30 * 1000
	REMEMBER_UPDATE_FREQ = 5 * 1000
)

// UpdateTTL runs an infinite while loop that updates internal ttl of all stored data objects
// relative to the real time that has passed.
// If the ttl of a some stored data reaches 0, the data is deleted
func (kademlia *Node) UpdateTTL() {
	lastUpdated := time.Now()
	for {
		// For each
		for dataHash, timeToLive := range kademlia.ttl {
			delta := time.Since(lastUpdated)
			kademlia.ttl[dataHash] -= int(delta.Milliseconds())

			if timeToLive <= 0 {
				kademlia.Delete(&dataHash)
				fmt.Println("Deleting hash", dataHash.String())
			}
		}
		lastUpdated = time.Now()
		// Sleeping to improve performance, no need to work all the time
		time.Sleep(1 * time.Second)
	}
}

// Remember runs an infinite while loop that sends refreshRPCs to all contact that is
// associated with some data that has been added via the put command (see cli.go)
// Runs local Refresh directly if one of the contacts are this node
func (network *Network) Remember() {
	if REMEMBER_UPDATE_FREQ >= TIME_TO_LIVE {
		fmt.Println("ERROR!  Update frequency of ttl refreshing is lower than the " +
			"system wide TTL parameter. No stored data will live for long ...")
	}
	for {
		// For each contact list associated to some data hash
		// (In other words, for each Store that this local node has initiated)
		for dataHash, contacts := range network.localNode.refreshContacts {
			for _, c := range contacts {
				if c.ID.Equals(network.localNode.routingTable.me.ID) {
					// Invoke local refresh directly, no reason to send RPCs to self
					//fmt.Println("Sending refresh to self")
					network.localNode.Refresh(&dataHash)
				} else {
					//fmt.Println("Sending refresh msg to", c.ID.String())
					network.refreshRPC(c, &dataHash)
				}
			}
		}
		time.Sleep(time.Duration(REMEMBER_UPDATE_FREQ) * time.Millisecond)
	}
}

// refreshRPC sends a REFRESH_DATA_TTL message to some kademlia node which will invoke the local
// node.Refresh function, effectively resetting the ttl for some hashed data so that the data
// won't be deleted
func (network *Network) refreshRPC(contact Contact, hash *KademliaID) {
	hostName := contact.Address
	service := hostName + ":" + KAD_PORT
	remoteAddr, err := network.ms_service.ResolveUDPAddr("udp",service)

	conn, err := network.ms_service.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Could not establish connection when sending refreshRPC to ", contact.ID.String(),"   ", contact.Address)
		return
	} else {
		// Message format:
		// SEND: [MSG TYPE, REQUESTER ID, REFRESH HASH]
		// REC: nothing

		msg := make([]byte, HEADER_LEN+ID_LEN+ID_LEN)
		msg[0] = REFRESH_DATA_TTL
		copy(msg[HEADER_LEN : HEADER_LEN+ID_LEN], network.localNode.routingTable.me.ID[:])
		copy(msg[HEADER_LEN+ID_LEN: HEADER_LEN+ID_LEN+ID_LEN], hash[:])
		conn.Write(msg)
	}
	conn.Close()
}