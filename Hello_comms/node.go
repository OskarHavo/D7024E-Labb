package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func CheckAndPrintError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func Send() {
	fmt.Println("Running sender!")

	for {
		var conn net.Conn
		var err error
		var port string
		for {
			fmt.Println("Enter port to ping: ")
			rawPort, inputErr := bufio.NewReader(os.Stdin).ReadString('\n')
			CheckAndPrintError(inputErr)
			port = strings.Split(rawPort, "\n")[0]

			fmt.Println("Establishing connection to localhost:" + port)
			conn, err = net.Dial("tcp", "localhost:" + port)
			if err == nil {
				break
			} else {
				fmt.Println("Could not dial localhost:" + port + ". Try again")
				fmt.Println(err)
			}
		}

		fmt.Println("Connection established! Sending ping msg.")
		conn.Write([]byte("This is a ping msg!" + "\n"))

		reply := make([]byte, 256)
		conn.Read(reply)

		fReply := strings.Split(string(reply), "\n")
		if fReply[0] == "Ack!" {
			fmt.Println("Received ack from localhost:" + port)
		} else {
			fmt.Println("Received unrecognized response from localhost: " + port)
		}

		fmt.Println("Closing connection ...")
		conn.Close()
	}
}


func Rec(port string) {
	fmt.Println("Running receiver!")

	fmt.Println("Listening to port " + port)
	conn, err := net.Listen("tcp", "localhost:" + port)
	CheckAndPrintError(err)

	for {
		c, err := conn.Accept()
		fmt.Println("Connection from " + c.RemoteAddr().String() + " accepted!")
		CheckAndPrintError(err)

		msg := make([]byte, 256)
		c.Read(msg)

		fMsg := strings.Split(string(msg), "\n")
		if fMsg[0] == "This is a ping msg!" {
			fmt.Println("Received ping from " + c.RemoteAddr().String())
		} else {
			fmt.Println("Received unrecognized msg from " + c.RemoteAddr().String())
		}

		fmt.Println("Sending Ack ...")
		reply := "Ack!"
		c.Write([]byte(reply + "\n"))
		c.Close()
	}
	conn.Close()
}

func main() {
	if len(os.Args[1:]) != 1 {
		fmt.Println("Expected 1 arguments. Got:", len(os.Args[1:]))
		os.Exit(-1)
	}

	args := os.Args[1:]
	go Rec(args[0])
	Send()
}