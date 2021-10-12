package main

import "sync"

const k = 20
const alpha = 3

// RoutingTable definition
// keeps a reference contact of me and an array of buckets
type RoutingTable struct {
	me      Contact
	buckets [ID_LEN * 8]*bucket
	bucketMutex sync.Mutex
}

// NewRoutingTable returns a new instance of a RoutingTable
func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{}
	for i := 0; i < ID_LEN*8; i++ {
		routingTable.buckets[i] = newBucket()
	}
	routingTable.me = me
	return routingTable
}

// AddContact add a new contact to the correct Bucket
func (routingTable *RoutingTable) AddContact(contact Contact) {
	routingTable.bucketMutex.Lock()
	defer routingTable.bucketMutex.Unlock()

	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.AddContact(contact)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	bucketIndex := routingTable.getBucketIndex(target)
	routingTable.bucketMutex.Lock()

	var candidates ContactCandidates
	defer routingTable.bucketMutex.Unlock()
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.GetContactsAndCalcDistances(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < ID_LEN*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactsAndCalcDistances(target))
		}
		if bucketIndex+i < ID_LEN*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.GetContactsAndCalcDistances(target))
		}
	}
	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(routingTable.me.ID)
	for i := 0; i < ID_LEN; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return ID_LEN*8 - 1
}

// KickTheBucket tries to remove an old node and put in a new one. The old contact will remain if it responds to a ping
func (routingTable *RoutingTable)KickTheBucket(contact *Contact, ping func(*Contact) bool) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]

	if bucket.Len() == k {
		element := bucket.Contains(contact)
		if element != nil {
			bucket.list.MoveToFront(element)
		} else {
			// Choose a node to sacrifice
			sacrifice := bucket.list.Back().Value.(Contact)

			if ping(&sacrifice) {

			} else {
				bucket.list.Remove(bucket.list.Back())
				bucket.AddContact(*contact)
			}
		}
	} else {
		bucket.AddContact(*contact)
	}
}
