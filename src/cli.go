package main // Funkar endast med "main"

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
		fmt.Println(help())
	default:
		fmt.Println("INVALID COMMAND, TYPE HELP")
	}
}
func handleDualInput(command string, value string, hashmap map[string]string) {

	switch command {
	case "put":
		outputHash := put(value, hashmap)
		fmt.Println(outputHash)
	case "get":
		outputNodeID, outputContent := get(value, hashmap)
		fmt.Println("NodeID: ", outputNodeID, "  Content: ", outputContent)

	default:
		fmt.Println("INVALID COMMAND, TYPE HELP")
	}
}

// Upload data of file downloaded
// Check if it can be uploaded
// if so, output the objects hash
func put(content string, hashmap map[string]string) string {

	hashedFileString := sha1Hash(content)
	_, exists := hashmap[hashedFileString] // Checks if value already exists

	if exists {
		return "Uploaded File Already Exists"
	} else {
		hashmap[hashedFileString] = content /// Adds the content and outputs the hash
		return hashedFileString
	}
}

// Take hash value as output
// Check if that exists in kademlia and download
// if so, output the contents of the objects and the node it was retrieved from.
func get(hashvalue string, hashmap map[string]string) (string, string) {
	nodeID := "000101010100101" // Temp value

	value, exists := hashmap[hashvalue] // Retrieve value from hashmap.

	if exists {
		return nodeID, value
	} else {
		return nodeID, ("Hashvalue Does Not Exist In The Hashmap")
	}
}

// Terminate node.
func exit() {
	os.Exit(1)
}

// Prints every command possible (return value due to testability)
func help() string {
	return ("Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
		"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
		"Exit -Terminates the node. " + "\n")
}

// Performs Sha-1 Hashing and encodes it into a String.
func sha1Hash(content string) string {
	// https://gobyexample.com/sha1-hashes
	h := sha1.New()
	h.Write([]byte(content))
	hashedFileBytes := h.Sum(nil)
	hashedFileString := hex.EncodeToString(hashedFileBytes) // Encode byte[] to string before entering it into the hashmap.
	return hashedFileString
}
