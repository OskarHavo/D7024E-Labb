package main

import (
	"net"
	"time"
)

type Message_service struct {
	use_fake bool
}

type Connection struct {
	conn *net.UDPConn
	use_fake bool
}

func (ms_service *Message_service) ListenUDP(udp string, addr *net.UDPAddr) (Connection, error) {
	if !ms_service.use_fake {
		conn, err := net.ListenUDP(udp,addr)
		return Connection{conn: conn,use_fake: ms_service.use_fake}, err
	} else {
		// TODO
		return Connection{conn: nil,use_fake: ms_service.use_fake}, nil
	}
}

func (ms_service *Message_service) ResolveUDPAddr(udp string, service string) (*net.UDPAddr, error){
	if !ms_service.use_fake {
		return net.ResolveUDPAddr(udp,service)
	} else {
		// TODO
		return nil, nil
	}
}

func (ms_service *Message_service) DialUDP(network string, laddr *net.UDPAddr, raddr *net.UDPAddr) (Connection, error) {
	if !ms_service.use_fake {
		conn, err := net.DialUDP(network,laddr,raddr)
		return Connection{conn: conn,use_fake: ms_service.use_fake}, err
	} else {
		return Connection{conn: nil,use_fake: ms_service.use_fake}, nil
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
		// TODO
		return 0,nil,nil
	}
}

func (connection *Connection) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	if !connection.use_fake {
		return connection.conn.WriteToUDP(b,addr)
	} else {
		// TODO
		return 0, nil
	}
}

func (connection *Connection) Write(b []byte) (int, error) {
	if !connection.use_fake {
		return connection.conn.Write(b)
	} else {
		// TODO
		return 0,nil
	}
}

func (connection *Connection) Close() {
	if !connection.use_fake {
		connection.conn.Close()
	} else {
		// TODO
	}
}