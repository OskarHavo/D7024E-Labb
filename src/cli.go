package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)
// Entrypoint
func main() {
	addrs,err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}
	var IP net.IP
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				IP = ipnet.IP
				fmt.Println(ipnet.IP.String() + "\n")
			}
		}
	}
	net := NewNetwork(&IP)
	fmt.Println("Started node with ID " + net.localNode.routingTable.me.ID.String())
	fmt.Println("Node has IP address " + IP.String())
	//go net.Listen()
	go net.HTTPlisten()

	for {
		fmt.Printf("\n Enter a command: ")
		rawInput, _ := bufio.NewReader(os.Stdin).ReadString('\n') // Takes rawinput from console.
		output := parseInput(rawInput, &net)
		fmt.Println("Returned output:\n" + output)
	}
}
// Parses the input and sends you to either the single/dual input handler.
func parseInput(input string, net *Network) string {
	var command string
	var value string

	stringinput := strings.Fields(input) //Splits the text into an array with each entry being a word

	if len(stringinput) == 0{
		fmt.Println("Blank input. Try again.")
	}
	// Single Input
	if len(stringinput) > 0 {
		command = stringinput[0]
		command = strings.ToLower(strings.Trim(command, " \r\n")) //Removes hidden \n etc, which makes string comparision impossible.
	}
	// Dual input
	if len(stringinput) > 1 { // Checks if you have 1 or 2 Commands and then runs the correct function accordingly.
		value = stringinput[1]
		// value = strings.ToLower(strings.Trim(value, " \r\n"))
		value = strings.Trim(value, " \r\n") // Will not make input lowercase (untested)
		return handleDualInput(command, value, net)
	} else {
		return handleSingleInput(command, 0)
	}
}
// Switch for all single input functions
func handleSingleInput(command string, testing int) string {
	switch command {
	case "exit":
		return exit(testing)
	case "help":
		return help()
	default:
		return "INVALID COMMAND, TYPE HELP"
	}
}
// Switch for all dual input functions
func handleDualInput(command string, value string, network *Network) string {
	switch command {
	case "put":
		return put(value, network)
	case "join":
		IP := net.ParseIP(value)[12:]
		ID := NewKademliaIDFromIP(&IP)
		network.Join(ID, value)
		return ""
	case "get":
		outputNodeID, outputContent := get(value, network)
		outputString := ("NodeID: " + outputNodeID + "  Content: " + outputContent)
		return outputString
	case "forget":
		// To-do
		return ""
	default:
		return "INVALID COMMAND, TYPE HELP"
	}
}

// Upload data of file downloaded. Check if it can be uploaded. If so, output the objects hash
func put(content string, net *Network) string {
	hashedFileString := NewKademliaIDFromData(content)
	net.Store([]byte(content),hashedFileString)
	return hashedFileString.String()
}

// Take hash value as output. Check if that exists in kademlia and download
// if so, output the contents of the objects and the node it was retrieved from.
func get(hashValue string, net *Network) (string, string) {
	hash := NewKademliaID(hashValue)
	data, nodes := net.DataLookup(hash)

	// TODO What ID should this be?
	if data != nil {
		return nodes[0].ID.String(),string(data)
	} else if len(nodes) > 0{
		return nodes[0].ID.String(), ("Hashvalue Does Not Exist In The Network")
	} else {
		return "[NULL]", ("Could not find node or data in the network")
	}
}

// Terminate node.
func exit(test int) string {
	if test != 0 {
		return "Exit (Test)"
	}
	os.Exit(1)
	return "Exit (Will not be reached)"
}

// Prints every command possible (return value due to testability)
func help() string {
	return "Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		    "Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		    "Exit -Terminates the node. " + "\n"
}