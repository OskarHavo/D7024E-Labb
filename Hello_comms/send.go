package main

import (
	"bufio"
	"fmt"
	"net"
)

func SendMsg() {

}

func RecMsg() {

}

func main() {
	fmt.Println("Running sender!")

	conn, err := net.Dial("tcp", "localhost:5001")
	CheckAndPrintError(err)

	defer conn.Close()

	msg := "Test msg from sender!"

	fmt.Fprintf(conn, msg + "\n")

	reply, _ := bufio.NewReader(conn).ReadString('\n')

	fmt.Println(string(reply))
}

func CheckAndPrintError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}