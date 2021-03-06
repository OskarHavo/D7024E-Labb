package main

import (
	"fmt"
	"sort"
)

// Contact definition
// stores the KademliaID, the ip address and the distance
type Contact struct {
	ID       *KademliaID
	Address  string
	distance *KademliaID
}

// NewContact returns a new instance of a Contact
func NewContact(id *KademliaID, address string) Contact {
	return Contact{id, address, nil}
}

// CalcDistance calculates the distance to the target and fills the contacts distance field
func (contact *Contact) CalcDistance(target *KademliaID) {
	contact.distance = contact.ID.CalcDistance(target)
}

// Less returns true if contact.distance < otherContact.distance
func (contact *Contact) Less(otherContact *Contact) bool {
	return contact.distance.Less(otherContact.distance)
}

// String returns a simple string representation of a Contact
func (contact *Contact) String() string {
	return fmt.Sprintf(`contact("%s", "%s")`, contact.ID, contact.Address)
}

// ContactCandidates definition
// Stores an array of Contacts
type ContactCandidates struct {
	contacts []Contact
}

// Append an array of Contacts to the ContactCandidates
func (candidates *ContactCandidates) Append(contacts []Contact) {
	candidates.contacts = append(candidates.contacts, contacts...)
}

// AppendContact appends a single contact rather than an array of contacts
func (candidates *ContactCandidates) AppendContact(contact Contact) {
	candidates.contacts = append(candidates.contacts, contact)
}

// GetContacts returns the first count number of Contacts. The complete list will be returned if there are fewer contacts.
func (candidates *ContactCandidates) GetContacts(count int) []Contact {
	if len(candidates.contacts) < count {
		return candidates.contacts
	} else {
		return candidates.contacts[:count]
	}
}

// Sort the Contacts in ContactCandidates
func (candidates *ContactCandidates) Sort() {
	sort.Sort(candidates)
}

// Len returns the length of the ContactCandidates
func (candidates *ContactCandidates) Len() int {
	return len(candidates.contacts)
}

// Swap the position of the Contacts at i and j
// WARNING does not check if either i or j is within range
func (candidates *ContactCandidates) Swap(i, j int) {
	candidates.contacts[i], candidates.contacts[j] = candidates.contacts[j], candidates.contacts[i]
}

// Less returns true if the Contact at index i is smaller than 
// the Contact at index j
func (candidates *ContactCandidates) Less(i, j int) bool {
	return candidates.contacts[i].Less(&candidates.contacts[j])
}

// Contains checks if the list of contacts already contains a node. This assumes that the
// candidates have been sorted and have calculated distances to the target node.
// No consideration is taken for unsorted contacts or contacts with improper distance.
func (candidates *ContactCandidates) Contains(contact *Contact) bool {
	for i := 0; i < candidates.Len(); i++ {
		if candidates.contacts[i].ID.Equals(contact.ID) {
			return true
		}
	}
	return false // If the candidates array is empty
}

// Remove removes a contact from an array if it exists, otherwise it does nothing
func (candidates *ContactCandidates) Remove(contact *Contact) {
	if !candidates.Contains(contact) {
		return
	}
	fmt.Println("Len of contact candidates before remove: ", candidates.Len())
	for i := 0; i < len(candidates.contacts); i++{
		if candidates.contacts[i].ID.Equals(contact.ID) {
			candidates.contacts = append(candidates.contacts[:i], candidates.contacts[i+1:]...)
			fmt.Println("Len of contact candidates after remove: ", candidates.Len())
			return
		}
	}
}