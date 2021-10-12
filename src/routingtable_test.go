package main

import (
	"testing"
)

func TestRoutingTable(t *testing.T) {
	result := []Contact{
		NewContact(NewKademliaID("0000000000000000000000000000000000000000"), ""),
		NewContact(NewKademliaID("000000000000000000000000000000000000000F"), ""),
		NewContact(NewKademliaID("00000000000000000000000000000000000000FF"), ""),
		NewContact(NewKademliaID("0000000000000000000000000000000000000FFF"), ""),
		NewContact(NewKademliaID("000000000000000000000000000000000000FFFF"), ""),
		NewContact(NewKademliaID("00000000000000000000000000000000000FFFFF"), ""),
		NewContact(NewKademliaID("0000000000000000000000000000000000FFFFFF"), "")}

	rt := NewRoutingTable(NewContact(NewKademliaID("0000000000000000000000000000000000FFFFFF"),""))

	rt.AddContact(result[5])
	rt.AddContact(result[4])
	rt.AddContact(result[2])
	rt.AddContact(result[3])
	rt.AddContact(result[1])
	rt.AddContact(result[6])
	rt.AddContact(result[0])

	t.Run("Closest", func(t *testing.T) {
		contacts := rt.FindClosestContacts(NewKademliaID("0000000000000000000000000000000000000000"), 20)

		for i := range contacts {
			if !result[i].ID.Equals(contacts[i].ID) {
				t.Errorf("Routing table test() = %v, want %v", contacts[i].ID.String(), result[i].ID.String())
			}
		}
	})
}
