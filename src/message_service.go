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

type Message_service struct {
	use_fake  bool
	home_addr *net.UDPAddr
	log_shannel chan string
}

type Connection struct {
	conn *net.UDPConn
	use_fake bool
	send_channel chan string
	send_IP string
	receive_channel chan string
	receive_IP string
}

func NewMessageService(use_fake bool, home_addr *net.UDPAddr) *Message_service {
	return &Message_service{use_fake: use_fake, home_addr: home_addr,log_shannel: make(chan string)}
}

var comm_mutex sync.Mutex
var global_map map[string] chan string = make(map[string] chan string)

func (ms_service *Message_service) ListenUDP(udp string, addr *net.UDPAddr) (Connection, error) {
	if !ms_service.use_fake {
		conn, err := net.ListenUDP(udp,addr)
		return Connection{conn: conn, use_fake: ms_service.use_fake}, err
	} else {
		receive_ID := ms_service.home_addr.IP.To4().String() + ":5001"
		// TODO
		comm_mutex.Lock()
		if global_map[receive_ID] == nil {
			global_map[receive_ID] = make(chan string)
		}
		comm_mutex.Unlock()
		//fmt.Println(receive_ID,"	: Listening to UDP")
		select {
		case s := <-global_map[receive_ID]:
			//fmt.Println(receive_ID,"	: Received connection from ", s)
			return Connection{conn: nil, use_fake: ms_service.use_fake, receive_channel: global_map[receive_ID],receive_IP: receive_ID, send_channel: global_map[s],send_IP: s}, nil
		case <-time.After(1*time.Second):
			return Connection{conn: nil, use_fake: ms_service.use_fake, receive_channel: nil}, errors.New("listened, but nobody answered")
		}
	}
}

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
		var port = rand.Intn(1000000)	// This is our temporary return address.
		home_addr := ms_service.home_addr.IP.To4().String() + ":" + strconv.FormatInt(int64(port), 10)
		//fmt.Println("-----------------------------", raddr)
		send_ID := raddr.IP.To4().String() + ":5001"
		//fmt.Println(home_addr, "	: Connecting to UDP address ", send_ID)
		comm_mutex.Lock()
		if global_map[send_ID] == nil {
			comm_mutex.Unlock()
			return Connection{conn: nil,use_fake: ms_service.use_fake}, errors.New("failed to dial udp")
		} else {
			//fmt.Println(home_addr, "	: Established connection at dialUDP to ",send_ID)
			global_map[home_addr] = make(chan string)
			comm_mutex.Unlock()
			global_map[send_ID] <- home_addr
			return Connection{conn: nil,use_fake: ms_service.use_fake,receive_channel: global_map[home_addr],receive_IP: home_addr,send_channel: global_map[send_ID],send_IP: send_ID},nil
		}
	}
}

func (connection *Connection) SetReadDeadline(time time.Time) {
	if !connection.use_fake {
		connection.conn.SetReadDeadline(time)
	} else {
		// TODO
	}
}

func (connection *Connection) ReadFromUDP(msg []byte) (n int, addr *net.UDPAddr,err error) {
	if !connection.use_fake {
		return connection.conn.ReadFromUDP(msg)
	} else {
		// TODO Add read timeout
		//fmt.Println(connection.receive_IP, "	: Starting to read from UDP channel ")
		copy(msg, []byte(<- connection.receive_channel))
		//fmt.Println(connection.receive_IP, "	: Received UDP packet. Packet size: ", len(msg))
		return len(msg),&net.UDPAddr{IP: net.ParseIP(connection.send_IP)},nil
	}
}

func (connection *Connection) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	if !connection.use_fake {
		return connection.conn.WriteToUDP(b,addr)
	} else {
		//fmt.Println(connection.receive_IP, "	: Starting to write to UDP channel", connection.send_IP)
		connection.send_channel <- string(b)
		//fmt.Println(connection.receive_IP, "	: Wrote ", len(b)," bytes to UDP successfully")
		return len(b), nil
	}
}

func (connection *Connection) Write(b []byte) (int, error) {
	if !connection.use_fake {
		return connection.conn.Write(b)
	} else {
		connection.send_channel <- string(b)
		return len(b),nil
	}
}

func (connection *Connection) Close() {
	if !connection.use_fake {
		connection.conn.Close()
	} else {
		/*
		i := 0
		for key,value := range global_map {
			if value == connection.send_channel || value == connection.receive_channel{
				i++
				delete(global_map, key)
			}
			if i == 2 {
				return
			}
		}
		 */
	}
}