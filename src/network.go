package d7024e

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Network struct {

}

func Listen(ip string, port int) {
	net.Listen("udp", ip + string(port))
}

func (network *Network) SendPingMessage(contact *Contact) {
	reply := make([]byte, 256)

	conn, err := net.Dial("udp", contact.Address + ":5001")
	fmt.Println("Connection established! Sending ping msg.")

	if err != nil {
		fmt.Println("Could not establish connection when pinging node " + contact.ID.String())
		// TODO: Tcp? Try again?
		return
	}

	start := time.Now()
	conn.Write([]byte("This is a ping msg!" + "\n"))
	conn.Read(reply)
	duration := time.Since(start)

	fReply := strings.Split(string(reply), "\n")
	if fReply[0] == "Ack!" {
		fmt.Println("Pinging node " + contact.ID.String() + " took " + strconv.FormatInt(duration.Milliseconds(), 10) + " ms")
	} else {
		fmt.Println("Received unrecognized response from node " + contact.ID.String() + " when pinged")
	}
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	conn, err = net.Dial("udp", contact.Address + ":5001")
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
