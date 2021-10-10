package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestRemoveSelfOrTail(t *testing.T) {
	// We need to test these cases:
	// Case 1. Remove self from slice correctly
	//    a) with removeTail
	//    b) without removeTail
	// Case 2. Not remove self from slice ...
	//    a) and remove tail
	//    b) and NOT remove tail

	mockReqIP := net.IP{192, 0, 0, 1}
	requesterID := NewKademliaIDFromIP(&mockReqIP)
	var b ContactCandidates

	// Case 1. Adding a bunch of ID's and one that matches requesterID
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000002"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000003"), "0"))
	b.AppendContact(NewContact(requesterID, "0"))
	oldSize := b.Len()
	result := removeSelfOrTail(requesterID, b.contacts, true)
	newSize := len(result)
	// requesterID should NOT be in the collection and collection should be size-1 from before
	if newSize != oldSize - 1 {
		t.Errorf("removeSelfOrTail did not remove an element when it should. Case 1a")
	}
	for _, e := range result {
		if *e.ID == *requesterID {
			t.Errorf("removeSelfOrTail did not remove requester ID properly. Case 1a")
		}
	}

	b = ContactCandidates{}
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000002"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000003"), "0"))
	b.AppendContact(NewContact(requesterID, ""))
	oldSize = b.Len()
	result = removeSelfOrTail(requesterID, b.contacts, false)
	newSize = len(result)
	// requesterID should NOT be in the collection and collection should be size-1 from before
	if newSize != oldSize - 1 {
		t.Errorf("removeSelfOrTail removed an element when it shouldn't. Case 1b")
	}
	for _, e := range result {
		if *e.ID == *requesterID {
			t.Errorf("removeSelfOrTail did not remove requester ID properly. Case 1b")
		}
	}

	// Case 2
	b = ContactCandidates{}
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000002"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000003"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000004"), "0"))
	oldSize = b.Len()
	result = removeSelfOrTail(requesterID, b.contacts, true)
	newSize = len(result)
	// last element should be gone and collection should be size-1 from before
	if newSize != oldSize - 1 {
		t.Errorf("removeSelfOrTail did not remove an element when it should. Case 2a.")
	}
	for _, e := range result {
		if *e.ID == *requesterID {
			t.Errorf("removeSelfOrTail did not remove requester ID properly. Case 2a")
		}
	}

	// Case 2
	b = ContactCandidates{}
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000002"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000003"), "0"))
	b.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000004"), "0"))
	oldSize = b.Len()
	result = removeSelfOrTail(requesterID, b.contacts, false)
	newSize = len(result)
	// last element should be gone and collection should be same size from before
	if newSize != oldSize {
		t.Errorf("removeSelfOrTail removed an element when it shouldn't. Case 2b.")
	}
	for _, e := range result {
		if *e.ID == *requesterID {
			t.Errorf("removeSelfOrTail did not remove requester ID properly. Case 2b")
		}
	}
}

func TestAddNewNodes(t *testing.T) {
	var v, u, n ContactCandidates
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000002"), "0"))
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000003"), "0"))

	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000004"), "0"))
	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000005"), "0"))
	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000006"), "0"))

	// Check for visited duplicates
	n.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	// Check for unvisited duplicates
	n.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000004"), "0"))
	// Check for unique new one
	n.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000007"), "0"))

	addNewNodes(&v, &u, n.GetContacts(3))

	foundCount := 0
	for _, e := range u.contacts {
		if *e.ID == *NewKademliaID("0000000000000000000000000000000000000001") {
			foundCount++
		}
	}
	if foundCount > 1 {
		t.Errorf("addNewNodes added a node that already exists in visited.")
	}

	foundCount = 0
	for _, e := range u.contacts {
		if *e.ID == *NewKademliaID("0000000000000000000000000000000000000004") {
			foundCount++
		}
	}
	if foundCount > 1 {
		t.Errorf("addNewNodes added a node that already exists in unvisited.")
	}

	foundCount = 0
	for _, e := range u.contacts {
		if *e.ID == *NewKademliaID("0000000000000000000000000000000000000007") {
			foundCount++
		}
	}
	if foundCount < 1 {
		t.Errorf("addNewNodes did not add a unique node for some reason.")
	}
}

func TestVisitedKClosest(t *testing.T) {
	k := 2
	var v, u ContactCandidates

	// Case 1: Have visited at least k contacts and all k contacts are the closest.
	// Should return true.
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	v.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000000")
	v.contacts[1].distance = NewKademliaID("0000000000000000000000000000000000000001")

	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	u.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000002")
	u.contacts[1].distance = NewKademliaID("0000000000000000000000000000000000000003")

	if !visitedKClosest(&u, &v, k) {
		t.Errorf("TestVisitedKClosest error case 1")
	}

	v = ContactCandidates{}
	u = ContactCandidates{}
	// Case 2: Have visited at least k contacts and all k EXCEPT 1 contact is the closest.
	// Should return false
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	v.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000000")
	v.contacts[1].distance = NewKademliaID("0000000000000000000000000000000000000002")

	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	u.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000001")
	u.contacts[1].distance = NewKademliaID("0000000000000000000000000000000000000003")

	if visitedKClosest(&u, &v, k) {
		t.Errorf("TestVisitedKClosest error case 2")
	}

	v = ContactCandidates{}
	u = ContactCandidates{}
	// Case 3: We have not visited k closest but no more unvisited to visit. Should return true
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	v.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000000")

	if !visitedKClosest(&u, &v, k) {
		t.Errorf("TestVisitedKClosest error case 3")
	}

	v = ContactCandidates{}
	u = ContactCandidates{}
	// Case 4: We have not visited k nodes yet, but there are more to visit. Should return false
	v.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	v.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000000")

	u.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0"))
	u.contacts[0].distance = NewKademliaID("0000000000000000000000000000000000000000")

	if visitedKClosest(&u, &v, k) {
		t.Errorf("TestVisitedKClosest error case 3")
	}
}

func TestDoWideSearch(t *testing.T) {
	closest := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0")

	var c []Contact
	c = append(c, NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	c = append(c, NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))

	// Case 1, closest is actually closest.
	// Should be WideSearch (round failed to return any node closer)
	closest.distance = NewKademliaID("0000000000000000000000000000000000000000")
	c[0].distance = NewKademliaID("0000000000000000000000000000000000000001")
	c[1].distance = NewKademliaID("0000000000000000000000000000000000000002")
	if !doWideSearch(&c, closest) {
		t.Errorf("TestDoWideSearch error case 1")
	}

	// Case 2, closest is actually NOT the closest.
	// Should not be WideSearch (round found a new closer node)
	closest.distance = NewKademliaID("0000000000000000000000000000000000000001")
	c[0].distance = NewKademliaID("0000000000000000000000000000000000000000")
	c[1].distance = NewKademliaID("0000000000000000000000000000000000000002")
	if doWideSearch(&c, closest) {
		t.Errorf("TestDoWideSearch error case 2")
	}
}

func TestHandleBucketReply(t *testing.T) {
	testNrContacts := 4
	s := HEADER_LEN+BUCKET_HEADER_LEN
	b := make([]byte, s+(ID_LEN+IP_LEN)*testNrContacts)
	b[HEADER_LEN] = byte(testNrContacts)
	IPs := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"}
	IDs := []KademliaID{*NewKademliaID("0000000000000000000000000000000000000000"),
		                *NewKademliaID("0000000000000000000000000000000000000001"),
		                *NewKademliaID("0000000000000000000000000000000000000002"),
		                *NewKademliaID("0000000000000000000000000000000000000003")}

	if len(IPs) != len(IDs) {
		fmt.Println("TestHandleBucketReply ERROR: IP and ID array are different sizes")
	}

	for i := 0; i < testNrContacts; i++ {
		copy(b[s+(ID_LEN+IP_LEN)*i : s+(ID_LEN+IP_LEN)*i+ID_LEN], IDs[i][:])

		copy(b[s+(ID_LEN+IP_LEN)*i+ID_LEN : s+(ID_LEN+IP_LEN)*i+ID_LEN+IP_LEN],
			net.ParseIP(IPs[i]).To4())
	}

	temp := handleBucketReply(&b)
	if temp.Len() != 4 {
		t.Errorf("handleBucketReply returned a bucket with incorrect size")
	}

	result := temp.GetContactsAndCalcDistances(NewKademliaID("0000000000000000000000000000000000000000"))
	// if each contact formed by IDs and IPs aren't in result, error
	for i := 0; i < testNrContacts; i++ {
		for j, c := range result {
			if IDs[i] == *c.ID && IPs[i] == c.Address {
				// found! Continue
				break
			} else {
				// not found ...
				if j == testNrContacts - 1 {
					// and we are at the last contact in result. Means the contact doesn't exist
					// and handleBucketReply doesn't extract it correctly. Error
					t.Errorf("handleBucketReply did not extract test element %d correctly", i)
				}
			}
		}
	}
}

func TestSetSearchSize(t *testing.T) {
	// case 1: we should visit k nodes if wideSearch and c contains at least k nodes
	var c ContactCandidates
	for i := 0; i < k; i++ {
		c.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	}
	if setSearchSize(true, &c) != k {
		t.Errorf("TestSetSearchSize case 1")
	}

	// case 2: we should visit alpha nodes if NOT wideSearch and c contains at least alpha nodes
	c = ContactCandidates{}
	for i := 0; i < alpha; i++ {
		c.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	}
	if setSearchSize(false, &c) != alpha {
		t.Errorf("TestSetSearchSize case 2")
	}

	// case 3a: we should visit c.Len() nodes if c.Len() < alpha <= k ...
	// (regardless of wideSearch)
	c = ContactCandidates{}
	for i := 0; i < alpha-1; i++ {
		c.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	}
	c3a := setSearchSize(false, &c)
	c3b := setSearchSize(true, &c)
	if c3a != c.Len() && c3b != c.Len() {
		t.Errorf("TestSetSearchSize case 3a")
	}

	// case 3b: ... OR alpha <= c.len <= k
	c = ContactCandidates{}
	for i := 0; i < k-1; i++ {
		c.AppendContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "0"))
	}
	c3c := setSearchSize(false, &c)
	c3d := setSearchSize(true, &c)
	if c3c != c.Len() && c3d != c.Len() {
		t.Errorf("TestSetSearchSize case 3b")
	}
}

func TestPostIterationProcessing(t *testing.T) {
	// We only have to test that postIterationProcessing moves contacts from unvisited to visited correctly.
	// Testing addNewNodes and doWideSearch is done separately (see this file)
	// We assume that postIterationProcessing will return the same value as doWideSearch
	// because it is IMPOSSIBLE TO MESS UP (1 LINE ASSIGNMENT)


	// We are checking that movement works for a
	// searchRange = min, searchRange = max and min < searchRange < max
	var v, u ContactCandidates
	var n []Contact
	c1 := NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "0")
	c2 := NewContact(NewKademliaID("0000000000000000000000000000000000000002"), "0")
	c3 := NewContact(NewKademliaID("0000000000000000000000000000000000000003"), "0")
	c4 := NewContact(NewKademliaID("0000000000000000000000000000000000000004"), "0")
	c1.distance = NewKademliaID("0000000000000000000000000000000000000000")
	c2.distance = NewKademliaID("0000000000000000000000000000000000000000")
	c3.distance = NewKademliaID("0000000000000000000000000000000000000000")
	c4.distance = NewKademliaID("0000000000000000000000000000000000000000")
	u.AppendContact(c1)
	u.AppendContact(c2)
	u.AppendContact(c3)

	s := 2
	// ID 1 and 2 should be moved
	postIterationProcessing(&v, &u, &n, s)
	if !v.Contains(&c1) || !v.Contains(&c2) || v.Contains(&c3) || u.Contains(&c1) || u.Contains(&c2) || !u.Contains(&c3){
		t.Errorf("postIterationProcessing did not move contacts properly from " +
			"unvisited to visited when searchRange = %d", s)
	}

	v = ContactCandidates{}
	u = ContactCandidates{}
	v.AppendContact(c4) // has to have at least 1 contact
	u.AppendContact(c1)
	u.AppendContact(c2)
	u.AppendContact(c3)
	s = 0
	// none should be moved
	postIterationProcessing(&v, &u, &n, s)
	if v.Len() > 1 {
		t.Errorf("postIterationProcessing did not move contacts properly from " +
			"unvisited to visited when searchRange = %d", s)
	}

	v = ContactCandidates{}
	u = ContactCandidates{}
	u.AppendContact(c1)
	u.AppendContact(c2)
	u.AppendContact(c3)
	s = 3
	// all should be moved
	postIterationProcessing(&v, &u, &n, s)
	if !v.Contains(&c1) || !v.Contains(&c2) || !v.Contains(&c3) || u.Contains(&c1) || u.Contains(&c2) || u.Contains(&c3){
		t.Errorf("postIterationProcessing did not move contacts properly from " +
			"unvisited to visited when searchRange = %d", s)
	}

	v = ContactCandidates{}
	u = ContactCandidates{}
	s = 100
	// all should be moved
	postIterationProcessing(&v, &u, &n, s)
	if v.Len() > 0 || u.Len() > 0 {
		t.Errorf("postIterationProcessing did not move contacts properly from " +
			"unvisited to visited when searchRange = %d", s)
	}
}

func TestNetwork_Listen(t *testing.T) {
	type fields struct {
		ip *net.IP
		ms_service *Message_service
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"", fields{&net.IP{},NewMessageService(false,&net.UDPAddr{})}},
		{"", fields{&net.IP{},NewMessageService(true,&net.UDPAddr{})}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network := NewNetwork(tt.fields.ip,tt.fields.ms_service)
			wait := make(chan bool)
			go func() {
				network.Listen()
				wait <- true
			}()
			time.Sleep(500*time.Millisecond)
			network.shutdown()
			select {
			case <- wait:
				return
			case <-time.After(1*time.Second):
				t.Errorf("Listen() = %v, want %v", network.running, false)
			}
		})
	}
}

func TestNetwork_Join(t *testing.T) {


	ip1 := net.ParseIP("0.0.0.0")
	ms1 := NewMessageService(true, &net.UDPAddr{IP: ip1})
	ip2 := net.ParseIP("0.0.0.1")
	ms2 := NewMessageService(true,&net.UDPAddr{IP: ip2})

	net1 := NewNetwork(&ip1,ms1)
	net2 := NewNetwork(&ip2,ms2)

	net1_chan := make(chan bool)
	go func() {
		net1.Listen()
		net1_chan <- true
	}()
	time.Sleep(50*time.Millisecond)
	error := net2.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")

	if error != nil {
		t.Errorf("Join() = %v, want %v", "Failed to join", "Succesful join")
	}
	net1.shutdown()
	<-net1_chan

	global_map = make(map[string] chan string)
	net2 = NewNetwork(&ip2,ms2)

	error = net2.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")
	if error == nil {
		t.Errorf("Join() = %v, want %v", "Succesful join","Failed to join")
	}

	global_map = make(map[string] chan string)
}

func TestNetwork_Store(t *testing.T) {

	global_map = make(map[string] chan string)

	ip1 := net.ParseIP("0.0.0.0")
	ms1 := NewMessageService(true, &net.UDPAddr{IP: ip1})
	ip2 := net.ParseIP("0.0.0.1")
	ms2 := NewMessageService(true,&net.UDPAddr{IP: ip2})

	net1 := NewNetwork(&ip1,ms1)
	net2 := NewNetwork(&ip2,ms2)

	data := []byte("Hello world!")
	net1.Store(data, NewKademliaIDFromData(string(data)))
	net1_chan := make(chan bool)
	go func() {
		net1.Listen()
		net1_chan <- true
	}()
	time.Sleep(50*time.Millisecond)
	error := net2.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")
	if error != nil {
		t.Errorf("Store() failed to create a connection. Check if join passed testing")
	}

	result,_ := net2.DataLookup(NewKademliaIDFromData(string(data)))

	if string(result[:12]) != string(data) {
		t.Errorf("Store() = %v, want %v", string(data),string(result))
	}

	data = []byte("Another text")
	net2.Store(data, NewKademliaIDFromData(string(data)))
	result,_ = net2.DataLookup(NewKademliaIDFromData(string(data)))

	if string(result[:12]) != string(data) {
		t.Errorf("Store() = %v, want %v", string(data),string(result))
	}
	net1.shutdown()
	<-net1_chan

	global_map = make(map[string] chan string)

}

func TestNetwork_NodeLookup(t *testing.T) {

	global_map = make(map[string] chan string)

	ip1 := net.ParseIP("0.0.0.0")
	ms1 := NewMessageService(true, &net.UDPAddr{IP: ip1})

	ip2 := net.ParseIP("0.0.0.1")
	ms2 := NewMessageService(true,&net.UDPAddr{IP: ip2})

	ip3 := net.ParseIP("0.0.0.2")
	ms3 := NewMessageService(true,&net.UDPAddr{IP: ip3})

	net1 := NewNetwork(&ip1,ms1)
	net2 := NewNetwork(&ip2,ms2)
	net3 := NewNetwork(&ip3,ms3)

	net1_chan := make(chan bool)
	go func() {
		net1.Listen()
		net1_chan <- true
	}()
	net2_chan := make(chan bool)
	go func() {
		net2.Listen()
		net2_chan <- true
	}()
	net3_chan := make(chan bool)
	go func() {
		net3.Listen()
		net3_chan <- true
	}()
	time.Sleep(50*time.Millisecond)
	error := net2.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")
	if error != nil {
		t.Errorf("NodeLookup() failed to create a connection. Check if join passed testing")
	}
	error = net3.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")
	if error != nil {
		t.Errorf("NodeLookup() failed to create a connection. Check if join passed testing")
	}
	contacts := net3.NodeLookup(NewKademliaIDFromIP(&ip1))

	if len(contacts) != 2 {
		t.Errorf("NodeLookup() = %v, want %v", len(contacts), 2)
	}

	net1.shutdown()
	net2.shutdown()
	net3.shutdown()
	<- net1_chan
	<- net2_chan
	<- net3_chan
}

func TestNetwork_DataLookup(t *testing.T) {

	global_map = make(map[string] chan string)

	ip1 := net.ParseIP("0.0.0.0")
	ms1 := NewMessageService(true, &net.UDPAddr{IP: ip1})

	ip2 := net.ParseIP("0.0.0.1")
	ms2 := NewMessageService(true,&net.UDPAddr{IP: ip2})

	ip3 := net.ParseIP("0.0.0.2")
	ms3 := NewMessageService(true,&net.UDPAddr{IP: ip3})

	net1 := NewNetwork(&ip1,ms1)
	net2 := NewNetwork(&ip2,ms2)
	net3 := NewNetwork(&ip3,ms3)

	net1_chan := make(chan bool)
	go func() {
		net1.Listen()
		net1_chan <- true
	}()
	net2_chan := make(chan bool)
	go func() {
		net2.Listen()
		net2_chan <- true
	}()
	net3_chan := make(chan bool)
	go func() {
		net3.Listen()
		net3_chan <- true
	}()
	time.Sleep(50*time.Millisecond)
	error := net2.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")
	if error != nil {
		t.Errorf("DataLookup() failed to create a connection. Check if join passed testing")
	}

	data := []byte("Hello world!")
	net1.Store(data, NewKademliaIDFromData(string(data)))

	error = net3.Join(NewKademliaIDFromIP(&ip1),"0.0.0.0")
	if error != nil {
		t.Errorf("DataLookup() failed to create a connection. Check if join passed testing")
	}
	result,contacts := net3.DataLookup(NewKademliaIDFromData(string(data)))
	if len(contacts) != 1 {
		t.Errorf("DataLookup() = %v, want %v", len(contacts), 1)
	} else if string(result[:12]) != string(data){
		t.Errorf("DataLookup() = %v, want %v", string(result), string(data))
	}

	data = []byte("Another hello!")
	result,contacts = net3.DataLookup(NewKademliaIDFromData(string(data)))
	if len(contacts) != 2 {
		t.Errorf("DataLookup() = %v, want %v", len(contacts), 2)
	}
	if result != nil{
		t.Errorf("DataLookup() = %v, want %v", string(result), string(data))
	}
	net1.shutdown()
	net2.shutdown()
	net3.shutdown()
	<- net1_chan
	<- net2_chan
	<- net3_chan

}