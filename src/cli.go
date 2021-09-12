package main //d7024e ? kompliererar ej

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func CheckAndPrintError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	hashmap := make(map[string]string) //temp for test

	for {
		// hashtable temp
		var command string
		var value string

		fmt.Printf("\n Enter a command: ")

		rawInput, inputErr := bufio.NewReader(os.Stdin).ReadString('\n') // Takes rawinput from console.
		CheckAndPrintError(inputErr)

		stringinput := strings.Fields(rawInput) //Splits the text into an array with each entry being a word

		command = stringinput[0]
		command = strings.ToLower(strings.Trim(command, " \r\n")) //Removes hidden \n etc, which makes string comparision impossible.

		if len(stringinput) > 1 { // Checks if you have 1 or 2 Commands and then runs the correct function accordingly.
			value = stringinput[1]
			value = strings.ToLower(strings.Trim(value, " \r\n"))
			handleDualInput(command, value, hashmap)
		} else {
			handleSingleInput(command)
		}
	}
}
func handleSingleInput(command string) {

	switch command {
	case "exit":
		exit()
	case "help":
		help()
	default:
		fmt.Println("INVALID COMMAND, TYPE HELP")
	}
}
func handleDualInput(command string, value string, hashmap map[string]string) {

	switch command {
	case "put":
		put(value, hashmap)
	case "get":
		get(value, hashmap)
	default:
		fmt.Println("INVALID COMMAND, TYPE HELP")
	}
}

// Upload data of file downloaded
// Check if it can be uploaded
// if so, output the objects hash
func put(content string, hashmap map[string]string) {
	// https://gobyexample.com/sha1-hashes
	h := sha1.New()
	h.Write([]byte(content))
	hashedFileBytes := h.Sum(nil)
	hashedFileString := hex.EncodeToString(hashedFileBytes) // Encode byte[] to string before entering it into the hashmap.

	_, exists := hashmap[hashedFileString] // Checks if value already exists

	if exists {
		fmt.Println("Uploaded File Already Exists")
	} else {
		hashmap[hashedFileString] = content /// Adds the content and outputs the hash
		fmt.Println(hashedFileString)
	}
}

// Take hash value as output
// Check if that exists in kademlia and download
// if so, output the contents of the objects and the node it was retrieved from.
func get(hashvalue string, hashmap map[string]string) {
	nodeID := "000101010100101" // Temp value

	value, exists := hashmap[hashvalue] // Retrieve value from hashmap.

	if exists {
		fmt.Println("Node: " + nodeID + "       Contains: " + value)
	} else {
		fmt.Println("Hashvalue Does Not Exist In The Hashmap")
	}
}

// Terminate node.
func exit() {
	os.Exit(0)
}

// Prints every command possible
func help() {
	fmt.Println("Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		"Exit -Terminates the node. " + "\n")
}
