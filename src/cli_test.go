package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestExit(t *testing.T) {
	// Test Exit
	output1 := exit(1)
	groundtruth1 := "Exit (Test)"
	if output1 != groundtruth1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output1, groundtruth1)
	} else {
		fmt.Println("EXIT - Test Exit = Passed") // -v must be added to go test for prints to appear.
	}
}
func TestHelp(t *testing.T) {
	// Functionality Test
	output1 := help()
	groundtruth1 := "Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		"Exit -Terminates the node. " + "\n"
	if output1 != groundtruth1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output1, groundtruth1)
	} else {
		fmt.Println("Help - Functionality Test = Passed") // -v must be added to go test for prints to appear.
	}
}
func TestHandleSingleInput(t *testing.T) {
	// Test Help
	output1 := handleSingleInput("help", 1)
	groundtruth1 := "Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		"Exit -Terminates the node. " + "\n"
	if output1 != groundtruth1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output1, groundtruth1)
	} else {
		fmt.Println("TestHandleSingleInput - Test Help = Passed") // -v must be added to go test for prints to appear.
	}

	// Test Default
	output2 := handleSingleInput("loremipsum", 1)
	groundtruth2 := "INVALID COMMAND, TYPE HELP"
	if output2 != groundtruth2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output2, groundtruth2)
	} else {
		fmt.Println("TestHandleSingleInput - Test Default = Passed") // -v must be added to go test for prints to appear.
	}

	// Test Exit
	output3 := handleSingleInput("exit", 1)
	groundtruth3 := "Exit (Test)"
	if output3 != groundtruth3 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output3, groundtruth3)
	} else {
		fmt.Println("TestHandleSingleInput - Test Exit = Passed") // -v must be added to go test for prints to appear.
	}

}
func TestParseInput(t *testing.T) {
	// Set Up
	addrs,_ := net.InterfaceAddrs()
	var testIP net.IP
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				testIP = ipnet.IP
			}
		}
	}
	net:= NewNetwork(&testIP, NewMessageService(false,nil))

	// Test no input
	output_0 := parseInput("", nil)
	groundTruth_0 := "Blank input. Try again.\n"
	if output_0 != groundTruth_0 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_0, groundTruth_0)
	} else {
		fmt.Println("TestParseInput - Test No Input = Passed") // -v must be added to go test for prints to appear.
	}

	// Test Single Input
	output_1 := parseInput("help", nil)
	groundTruth_1 := "Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		"Exit -Terminates the node. " + "\n"
	if output_1 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1, groundTruth_1)
	} else {
		fmt.Println("TestParseInput - Test Single Input = Passed") // -v must be added to go test for prints to appear.
	}
	// Test Dual Input
	output_2 := parseInput("put test", &net)
	groundTruth_2 := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	if output_2 != groundTruth_2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_2, groundTruth_2)
	} else {
		fmt.Println("TestParseInput - Test Dual Input = Passed") // -v must be added to go test for prints to appear.
	}
}

func TestHandleDualInput(t *testing.T) {
	// Set Up
	addrs,_ := net.InterfaceAddrs()
	var testIP net.IP
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				testIP = ipnet.IP
			}
		}
	}
	network:= NewNetwork(&testIP,NewMessageService(false,nil))
	// Test join
	{
		ip1 := net.ParseIP("0.0.0.0")
		ms1 := NewMessageService(true, &net.UDPAddr{IP: ip1})
		ip2 := net.ParseIP("0.0.0.1")
		ms2 := NewMessageService(true, &net.UDPAddr{IP: ip2})

		net1 := NewNetwork(&ip1, ms1)
		net2 := NewNetwork(&ip2, ms2)

		net1_chan := make(chan bool)
		go func() {
			net1.Listen()
			net1_chan <- true
		}()
		time.Sleep(50*time.Millisecond)

		output := handleDualInput("join","0.0.0.0",&net2)
		groundTruth := ""
		if output != groundTruth {
			t.Errorf("Answer was incorrect, got: %s, want: %s.", output, groundTruth)
		} else {
			fmt.Println("TestHandleDualInput - Test join = Passed") // -v must be added to go test for prints to appear.
		}

		net1.shutdown()
		<- net1_chan
	}
	// Test join with error
	{
		ip2 := net.ParseIP("0.0.0.1")
		ms2 := NewMessageService(true, &net.UDPAddr{IP: ip2})
		net2 := NewNetwork(&ip2, ms2)

		output := handleDualInput("join","0.0.0.0",&net2)
		groundTruth := "could not join network node"
		if output != groundTruth {
			t.Errorf("Answer was incorrect, got: %s, want: %s.", output, groundTruth)
		} else {
			fmt.Println("TestHandleDualInput - Test join = Passed") // -v must be added to go test for prints to appear.
		}

	}

	// Test Put
	output_1 := handleDualInput("put", "test", &network)
	groundTruth_1 := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	if output_1 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1, groundTruth_1)
	} else {
		fmt.Println("TestHandleDualInput - Test Put = Passed") // -v must be added to go test for prints to appear.
	}
	// Test Default
	output_2 := handleDualInput("lorem", "ipsum", &network)
	groundTruth_2 := "INVALID COMMAND, TYPE HELP"
	if output_2 != groundTruth_2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_2, groundTruth_2)
	} else {
		fmt.Println("TestHandleDualInput - Test Default = Passed") // -v must be added to go test for prints to appear.
	}
	// Test Get
	inputString := "test"
	put(inputString, &network)
	output_3 := handleDualInput("get", NewKademliaIDFromData(inputString).String(), &network)
	groundTruth_3 := "NodeID: "+ network.localNode.routingTable.me.ID.String() +"  Content: test"
	if output_3 != groundTruth_3 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_3, groundTruth_3)
	} else {
		fmt.Println("TestHandleDualInput - Test Put = Passed") // -v must be added to go test for prints to appear.
	}
}


func TestPut(t *testing.T) {
	// Set Up
	addrs,_ := net.InterfaceAddrs()
	var testIP net.IP
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				testIP = ipnet.IP
			}
		}
	}
	net:= NewNetwork(&testIP,NewMessageService(false,nil))

	// Test Good Input
	output_1 := put("testing", &net)
	groundTruth_1 := "dc724af18fbdd4e59189f5fe768a5f8311527050"
	if output_1 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1, groundTruth_1)
	} else {
		fmt.Println("PUT - Good Input = Passed") // -v must be added to go test for prints to appear.
	}
}

func TestGet(t *testing.T) {
	// Set Up
	addrs,_ := net.InterfaceAddrs()
	var testIP net.IP
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				testIP = ipnet.IP
			}
		}
	}
	net:= NewNetwork(&testIP,NewMessageService(false,nil))

	// Test Find Valid Input
	inputString := "test"
	input_1 := put(inputString, &net)
	_, output_1_2 := get(input_1, &net)
	groundTruth_1 := inputString
	if output_1_2 != groundTruth_1 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_1_2, groundTruth_1)
	} else {
		fmt.Println("GET - Find Valid Input = Passed")
	}

	// Test Find Null Input
	inputHash := "0000000000000000000000000000000000000000"
	_, output_2_2 := get(inputHash, &net)
	groundTruth_2 := "Could not find node or data in the network"
	if output_2_2 != groundTruth_2 {
		t.Errorf("Answer was incorrect, got: %s, want: %s.", output_2_2, groundTruth_2)
	} else {
		fmt.Println("GET - Find Null Input = Passed")
	}
}

