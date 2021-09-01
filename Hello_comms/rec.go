package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	fmt.Println("Running receiver!")

	fmt.Println("Listening to port ...")
	con, err := net.Listen("tcp", "localhost:5001")
	CheckAndPrintError(err)

	defer con.Close()

	c, err := con.Accept()
	fmt.Println("Connection accepted!")
	CheckAndPrintError(err)
	fmt.Println("Handling connection ...")
	HandleConnection(c)
}

func HandleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	msg, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(msg)

	reply := "Response msg from receiver!"
	c.Write([]byte(reply))
	c.Close()
}

func CheckAndPrintError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
