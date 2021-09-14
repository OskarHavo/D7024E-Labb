package main

import (
	"fmt"
	"testing"
)

// Files must be xxx_test.go
// Functions must be TestXxxx(t *testing.T)
// Then --> go test cli_test.go cli.go -v -cover

func TestPut(t *testing.T) {
	// Set Up
	hashmap := make(map[string]string) //temp for test

	// Test Good Input
	output_1 := put("test", hashmap)
	groundTruth_1 := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	if output_1 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1, groundTruth_1)
	} else {
		fmt.Println("PUT - Good Input = Passed") // -v must be added to go test for prints to appear.
	}
	// Test Already Used Input
	output_2 := put("test", hashmap)
	groundTruth_2 := "Uploaded File Already Exists"
	if output_2 != groundTruth_2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_2, groundTruth_2)
	} else {
		fmt.Println("PUT - Already Used Input = Passed") // -v must be added to go test for prints to appear.
	}
}
func TestGet(t *testing.T) {
	// Set Up
	hashmap := make(map[string]string) //temp for test

	// Test Find Valid Input
	inputString := "test"
	input_1 := put(inputString, hashmap)
	_, output_1_2 := get(input_1, hashmap)
	groundTruth_1 := inputString
	if output_1_2 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1_2, groundTruth_1)
	} else {
		fmt.Println("GET - Find Valid Input = Passed")
	}

	// Test Find Invalid Input
	inputHash := "loremipsum"
	_, output_2_2 := get(inputHash, hashmap)
	groundTruth_2 := "Hashvalue Does Not Exist In The Hashmap"
	if output_2_2 != groundTruth_2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_2_2, groundTruth_2)
	} else {
		fmt.Println("GET - Find Invalid Input = Passed")
	}
}
func TestExit(t *testing.T) {
	// Difficult to do without ALOT of work arounds.
	t.Skipped()
}

func TestHelp(t *testing.T) {
	// Functionality Test
	output_1 := help()
	groundTruth_1 := "Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		"Exit -Terminates the node. " + "\n"
	if output_1 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1, groundTruth_1)
	} else {
		fmt.Println("Help - Functionality Test = Passed") // -v must be added to go test for prints to appear.
	}
}

func TestSha1Hash(t *testing.T) {
	// Test Good Input
	output_1 := sha1Hash("test")
	groundTruth_1 := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	if output_1 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1, groundTruth_1)
	} else {
		fmt.Println("Sha1Hash - Good Input = Passed") // -v must be added to go test for prints to appear.
	}
}
