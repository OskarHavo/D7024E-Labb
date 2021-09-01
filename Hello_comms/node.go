package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func HandleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	msg, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	fMsg := strings.TrimSuffix(msg, "\n")
	fmt.Println("Message \"" + fMsg + "\" was received successfully!. Sending ack ...")

	reply := "Ack!"
	c.Write([]byte(reply))
	c.Close()
}

func CheckAndPrintError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func Rec(port string) {
	for {
		fmt.Println("Running receiver!")

		fmt.Println("Listening to port " + port)
		con, err := net.Listen("tcp", "localhost:" + port)
		CheckAndPrintError(err)

		c, err := con.Accept()
		fmt.Println("Connection accepted!")
		CheckAndPrintError(err)
		fmt.Println("Handling connection ...")
		HandleConnection(c)

		con.Close()
	}
}

func Send(port string) {
	for {
		fmt.Println("Running sender!")

		var conn net.Conn
		var err error
		conn, err = net.Dial("tcp", "localhost:" + port)
		for err != nil {
			conn, err = net.Dial("tcp", "localhost:" + port)
		}
		CheckAndPrintError(err)

		//msg := "Test msg from sender!"
		// Get msg to send from keyboard input
		fmt.Println("Enter msg to send: ")
		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
		CheckAndPrintError(err)
		fmt.Fprintf(conn, msg + "\n") // Send msg

		reply, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Println("Received response: " + reply)

		conn.Close()
	}
}

func main() {
	if len(os.Args[1:]) != 2 {
		fmt.Println("Expected 2 arguments. Got:", len(os.Args[1:]))
		os.Exit(-1)
	}

	args := os.Args[1:]
	go Rec(args[0])
	Send(args[1])
}