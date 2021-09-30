package main

import (
	"fmt"
	"testing"
)

// Tests so that a new node is actually of type Node.
func TestNewNode(t *testing.T) {
	// Setup
	testId :=  NewKademliaIDFromData("test")
	testContact := NewContact(testId,"0.0.0.0")
	testNode := NewNode(testContact)

	output1 := testIfNode(testNode)
	groundtruth1 := true
	if output1 != groundtruth1 {
		t.Errorf("Type from newNode is not Node")
	} else {
		fmt.Println("TestNewNode = Passed") // -v must be added to go test for prints to appear.
	}
}
func testIfNode(t interface{}) bool{
	switch t.(type){
	case Node:
		return true
	default:
		return false
	}
}
// Tests so you can Store data, and then Lookup this data.
func TestLookupData(t *testing.T) {
	// Setup
	testString:= "hello"
	testStringAsByteArray := []byte(testString)
	testId :=  NewKademliaIDFromData(testString)
	testContact := NewContact(testId,"0.0.0.0")
	testNode := NewNode(testContact)

	// Check if Store adds something to Node
	testNode.Store(testStringAsByteArray, testId)

	output1 := len(testNode.storage)
	groundtruth1 := 1
	if output1 != groundtruth1 {
		t.Errorf("Answer was incorrect, got: %d, want: %d.", output1, groundtruth1)
	} else {
		fmt.Println("TestLookupData - Store = Passed") // -v must be added to go test for prints to appear.
	}

	// Check if the ID will return testString "hello"
	output2 := string(testNode.LookupData(testId))
	groundtruth2 := "hello"
	if output2 != groundtruth2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output2, groundtruth2)
	} else {
		fmt.Println("TestLookupData - TestLookupData = Passed") // -v must be added to go test for prints to appear.
	}
}