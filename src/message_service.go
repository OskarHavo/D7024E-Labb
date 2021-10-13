package main

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// This is our solution for simulated network UP messages.
// The message service can eithe be configured to create real network connections or fake ones.
type Message_service struct {
	use_fake  bool
	home_addr *net.UDPAddr
	log_shannel chan string
}

// This is a simulated network connection.
type Connection struct {
	conn *net.UDPConn
	use_fake bool
	send_channel chan string
	send_IP string
	receive_channel chan string
	receive_IP string
	readDeadline time.Time
	hasDeadline bool
}

func NewMessageService(use_fake bool, home_addr *net.UDPAddr) *Message_service {
	return &Message_service{use_fake: use_fake, home_addr: home_addr,log_shannel: make(chan string)}
}

// This is a global data structure that keeps track of all available connections.
// Kademlia nodes that start to listen will advertise their connection channel here
var comm_mutex sync.Mutex
var global_map map[string] chan string = make(map[string] chan string)

func (ms_service *Message_service) ListenUDP(udp string, addr *net.UDPAddr) (Connection, error) {
	if !ms_service.use_fake {
		conn, err := net.ListenUDP(udp,addr)
		return Connection{conn: conn, use_fake: ms_service.use_fake}, err
	} else {
		// Start to listen for a fake UDP connection

		receive_ID := ms_service.home_addr.IP.To4().String() + ":5001"	// Create a fake port and IP combo
		comm_mutex.Lock()
		if global_map[receive_ID] == nil {
			global_map[receive_ID] = make(chan string)
		}
		comm_mutex.Unlock()
		select {
		case s := <-global_map[receive_ID]:	// Start to listen for 1 second before timeout.
			return Connection{conn: nil, use_fake: ms_service.use_fake, receive_channel: global_map[receive_ID],receive_IP: receive_ID, send_channel: global_map[s],send_IP: s}, nil
		case <-time.After(1*time.Second):
			return Connection{conn: nil, use_fake: ms_service.use_fake, receive_channel: nil}, errors.New("listened, but nobody answered")
		}
	}
}

// Create an IP address from a string. There was a problem with fake IPs but I can't rememever what. This fixes that.
func (ms_service *Message_service) ResolveUDPAddr(udp string, service string) (*net.UDPAddr, error){
	if ms_service.use_fake {
		serv := strings.Split(service, ":")
		port,_ := strconv.Atoi(serv[1])
		return &net.UDPAddr{IP: net.ParseIP(serv[0]),Port: port}, nil
	} else {
		return net.ResolveUDPAddr(udp,service)
	}
}

func (ms_service *Message_service) DialUDP(network string, laddr *net.UDPAddr, raddr *net.UDPAddr) (Connection, error) {
	if !ms_service.use_fake {
		conn, err := net.DialUDP(network,laddr,raddr)
		return Connection{conn: conn,use_fake: ms_service.use_fake}, err
	} else {
		// Dial to a fake UDP connection.

		// Set up our own channel for incoming messages.
		var port = rand.Intn(1000000)	// This is our temporary return address.
		home_addr := ms_service.home_addr.IP.To4().String() + ":" + strconv.FormatInt(int64(port), 10)
		send_ID := raddr.IP.To4().String() + ":5001"

		comm_mutex.Lock()
		if global_map[send_ID] == nil {
			comm_mutex.Unlock()
			return Connection{conn: nil,use_fake: ms_service.use_fake}, errors.New("failed to dial udp")
		} else {
			global_map[home_addr] = make(chan string)
			comm_mutex.Unlock()

			select {
			case global_map[send_ID] <- home_addr:	// Try call the other node.
				return Connection{conn: nil, use_fake: ms_service.use_fake, receive_channel: global_map[home_addr], receive_IP: home_addr, send_channel: global_map[send_ID], send_IP: send_ID}, nil
			case <-time.After(1 * time.Second):
				return Connection{conn: nil, use_fake: ms_service.use_fake, receive_channel: nil, receive_IP: home_addr, send_channel: nil, send_IP: send_ID}, errors.New("failed to dial udp")

			}
		}
	}
}

func (connection *Connection) SetReadDeadline(t time.Time) {
	if !connection.use_fake {
		connection.conn.SetReadDeadline(t)
	} else {
		connection.readDeadline = t
		connection.hasDeadline = true
	}
}

func (connection *Connection) ReadFromUDP(msg []byte) (n int, addr *net.UDPAddr,err error) {
	if !connection.use_fake {
		return connection.conn.ReadFromUDP(msg)
	} else {

		if connection.hasDeadline {
			connection.hasDeadline = false

			// Read with a timeout.
			select {
			case data := <- connection.receive_channel:
				copy(msg, []byte(data))
				return len(msg),&net.UDPAddr{IP: net.ParseIP(connection.send_IP)},nil
			case <- time.After(connection.readDeadline.Sub(time.Now())):
				return 0,&net.UDPAddr{IP: net.ParseIP(connection.send_IP)},errors.New("Could not read from UDP")
			}
		} else {
			// Don't set a read timeout.
			data := <- connection.receive_channel
			copy(msg, []byte(data))
			return len(msg),&net.UDPAddr{IP: net.ParseIP(connection.send_IP)},nil
		}
	}
}

func (connection *Connection) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	// TODO We don't have a write deadline. One second is the default right now
	if !connection.use_fake {
		return connection.conn.WriteToUDP(b,addr)
	} else {
		select {
		case connection.send_channel <- string(b):
			return len(b),nil
		case <- time.After(1*time.Second):
			return 0, errors.New("Could not write to UDP")
		}
	}
}

func (connection *Connection) Write(b []byte) (int, error) {
	// TODO We don't have a write deadline. One second is the default right now
	if !connection.use_fake {
		return connection.conn.Write(b)
	} else {
		select {
		case connection.send_channel <- string(b):
			return len(b),nil
		case <- time.After(1*time.Second):
			return 0, errors.New("Could not write to UDP")
		}
	}
}

// This only works for real connections.
func (connection *Connection) Close() {
	if !connection.use_fake {
		connection.conn.Close()
	}
}