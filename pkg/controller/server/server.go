package server

import (
	"dns-proxy/pkg/domain/proxy"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type DNSServer struct {
	proxySvc          proxy.Service
	port              int
	host              string
	maxPoolConnection int
}

func New(proxy proxy.Service, port int, host string, maxPoolConnection int) *DNSServer {
	return &DNSServer{
		proxySvc:          proxy,
		port:              port,
		host:              host,
		maxPoolConnection: maxPoolConnection,
	}

}

func (d *DNSServer) StartTCPServer() {
	portStr := strconv.Itoa(d.port)
	fmt.Println("Listening TCP DNS Proxy on PORT " + portStr)
	ln, err := net.Listen("tcp", d.host+":"+portStr)
	if err != nil {
		log.Println("Error creating listener")
		panic(err)
	}
	var conns uint64
	for {
		if conns <= uint64(d.maxPoolConnection-1) {
			conn, err := ln.Accept()
			atomic.AddUint64(&conns, 1)
			if err != nil {
				log.Println("Error creating connection")
				panic(err)
			}
			go tcpHandler(&conn, d.proxySvc, &conns)
		}
	}
}

func (d *DNSServer) StartUDPServer() {
	portStr := strconv.Itoa(d.port)
	//TODO: Change this for a logger
	fmt.Println("Listening UDP DNS Proxy on PORT " + portStr)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: d.port})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	udpHandler(conn, d.proxySvc)
}

func tcpHandler(conn *net.Conn, p proxy.Service, conns *uint64) error {
	defer (*conn).Close()
	var unsolvedMsg [2024]byte
	n, err := (*conn).Read(unsolvedMsg[:])
	if err != nil {
		log.Println("Failed to read from connection.")
		return errors.New("Failed to read from connection.")
	}
	solvedMsg, proxyErr := p.SolveTCP(unsolvedMsg[:n])
	if proxyErr != nil {
		fmt.Printf("Error solving message: %v \n", proxyErr)
	}
	(*conn).Write(solvedMsg)
	atomic.AddUint64(conns, ^uint64(0))
	return nil
}

func udpHandler(conn *net.UDPConn, p proxy.Service) {
	for {
		unsolvedMsg := make([]byte, 2048)
		n, addr, err := conn.ReadFromUDP(unsolvedMsg)
		if err != nil {
			log.Println("Failed to read from connection.")
		}

		solvedMsg, err := p.SolveUDP(unsolvedMsg[:n])
		if err != nil {
			fmt.Printf("Error solving message: %v \n", err)
		}

		_, err = conn.WriteToUDP(solvedMsg[:], addr)
		if err != nil {
			fmt.Println(err)
		}
	}
}
