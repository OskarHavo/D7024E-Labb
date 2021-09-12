package main //d7024e ? kompliererar ej

import (
	"bufio"
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
	for {
		var command string
		var value string

		fmt.Printf("\n Enter a command: ")

		rawInput, inputErr := bufio.NewReader(os.Stdin).ReadString('\n') // Takes rawinput from console.
		CheckAndPrintError(inputErr)

		stringinput := strings.Fields(rawInput) //Splits the text into an array with each entry being a word

		//fmt.Println(stringinput, "    1:   ", stringinput[0], "    2: ", stringinput[1])
		command = stringinput[0]
		value = stringinput[1]

		command = strings.ToLower(strings.Trim(command, " \r\n")) //Removes hidden \n etc, which makes string comparision impossible.
		value = strings.ToLower(strings.Trim(value, " \r\n"))

		//fmt.Println(command + " and " + value)
		handleInput(command, value)
	}

}
func handleInput(command string, value string) {

	switch command {
	case "put":
		put(value)
	case "get":
		get(value)
	case "exit":
		exit()
	case "help":
		fmt.Println("Put - Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully." + "\n" +
			"Get - Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully. " + "\n" +
			"Exit -Terminates the node. " + "\n")
	default:
		fmt.Println("INVALID COMMAND, TYPE HELP")
	}

}
func put(file string) string {
	// Upload data of file downloaded
	// Check if it can be uploaded
	// if so, output the objects hash

	fmt.Println("put called")
	if file == "test" {
		fmt.Println("test works")
	}

	return "Placeholder Hash : 34873874387473847328743478"

}
func get(hashvalue string) (string, string) {
	// Take hash value as output
	// Check if that exists in kademlia and download
	// if so, output the contents of the objects and the node it was retrieved from.

	fmt.Println("get called")

	contents := "Virus"
	nodeID := "000101010100101"
	return ("Node ID " + nodeID), ("contains: " + contents)

}
func exit() {
	// Terminate node.
	fmt.Println("exit called")

}
