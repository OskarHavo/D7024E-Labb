package sprint0

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

func Send(port string) {
	fmt.Println("[SEND] Running sender!")

	for {
		var conn net.Conn
		var err error
		var ip string
		for {
			fmt.Println("[REC] Enter [ip-address] to ping: ")
			rawAP, inputErr := bufio.NewReader(os.Stdin).ReadString('\n')
			CheckAndPrintError(inputErr)
			ip = strings.Split(rawAP, "\n")[0]

			fmt.Println("[REC] Establishing connection to " + ip)
			conn, err = net.Dial("tcp", ip + ":" + port)
			if err == nil {
				break
			} else {
				fmt.Println("[REC] Could not dial " + ip + ". Try again")
				fmt.Println(err)
			}
		}

		fmt.Println("[REC] Connection established! Sending ping msg.")
		conn.Write([]byte("This is a ping msg!" + "\n"))

		reply := make([]byte, 256)
		conn.Read(reply)

		fReply := strings.Split(string(reply), "\n")
		if fReply[0] == "Ack!" {
			fmt.Println("[REC] Received ack from " + ip)
		} else {
			fmt.Println("[REC] Received unrecognized response from " + ip)
		}

		fmt.Println("[REC] Closing connection ...")
		conn.Close()
	}
}


func Rec(port string) {
	fmt.Println("[REC] Running receiver!")

	fmt.Println("[SEND] Listening to port " + port)
	conn, err := net.Listen("tcp", ":" + port)
	CheckAndPrintError(err)

	for {
		c, err := conn.Accept()
		fmt.Println("[SEND] Connection from " + c.RemoteAddr().String() + " accepted!")
		CheckAndPrintError(err)

		msg := make([]byte, 256)
		c.Read(msg)

		fMsg := strings.Split(string(msg), "\n")
		if fMsg[0] == "This is a ping msg!" {
			fmt.Println("[SEND] Received ping from " + c.RemoteAddr().String())
		} else {
			fmt.Println("[SEND] Received unrecognized msg from " + c.RemoteAddr().String())
		}

		fmt.Println("[SEND] Sending Ack ...")
		reply := "Ack!"
		c.Write([]byte(reply + "\n"))
		c.Close()
	}
	conn.Close()
}

func main() {
	port := "5001"
	go Rec(port)
	Send(port)
}