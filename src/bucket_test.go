package main

import (
	"fmt"
	"reflect"
	"testing"
)

// Tests so that a new Bucket is actually of type bucket.
func TestNewBucket(t *testing.T) {
	// Test newBucket
	testBucket := newBucket()
	output1 := testIfBucket(testBucket)
	groundtruth1 := true
	if output1 != groundtruth1 {
		t.Errorf("Type from newBucket is not Bucket")
	} else {
		fmt.Println("TestNewBucket = Passed") // -v must be added to go test for prints to appear.
	}
}
func testIfBucket(t interface{}) bool{
	switch t.(type){
		case *bucket:
			return true
		default:
			return false
	}
}
// Tests so that adding a contact increases the length of the bucket.
// Also tests Len()
func TestAddContact(t *testing.T) {
	// Setup
	testBucket := newBucket()
	testId :=  NewKademliaIDFromData("test")
	testContact := NewContact(testId,"0.0.0.0")

	// Check if adding increments the length
	testBucket.AddContact(testContact)
	output1 := testBucket.Len()  // 1 = 1
	groundtruth1 := 1
	if output1 != groundtruth1 {
		t.Errorf("Answer was incorrect, got: %d, want: %d.", output1, groundtruth1)
	} else {
		fmt.Println("TestAddContact = Passed") // -v must be added to go test for prints to appear.
	}
}

// Tests so that contains returns elements from the bucket's list.
func TestContains(t *testing.T) {
	// Setup
	testBucket := newBucket()
	testId :=  NewKademliaIDFromData("test")
	testContact := NewContact(testId,"0.0.0.0")

	// Check contains returns the element from the bucket's list.
	testBucket.AddContact(testContact)

	output1 := testBucket.Contains(&testContact)
	groundtruth1 := testBucket.list.Front()
	if output1 != groundtruth1 {
		t.Errorf("Element is not returned correctly")
	} else {
		fmt.Println("TestContains = Passed") // -v must be added to go test for prints to appear.
	}
}

// Checks if the returned array's first value's distance has been added.
func TestGetContactsAndCalcDistances(t *testing.T) {
	// Setup
	testBucket := newBucket()
	testId :=  NewKademliaIDFromData("test")
	testContact := NewContact(testId,"0.0.0.0")

	// Check contains returns the element from the bucket's list.
	testBucket.AddContact(testContact)
	input := testBucket.GetContactsAndCalcDistances(testId)[0].distance
	output1 := reflect.ValueOf(input).IsNil()
	groundtruth1 := false
	if output1 != groundtruth1 {
		t.Errorf("The calculated distance is nil")
	} else {
		fmt.Println("TestGetContactsAndCalcDistances = Passed") // -v must be added to go test for prints to appear.
	}
}