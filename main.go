package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	conn_type = "tcp"
	host      = "0.0.0.0"
	port      = "8080"
)

func RemoveConn(slice *[]net.Conn, conn net.Conn) *[]net.Conn {
	var indexToRemove int
	for i, s := range *slice {
		if s == conn {
			fmt.Println("tá certo")
			indexToRemove = i
		}
	}
	slice2 := *slice
	slice2 = append(slice2[:indexToRemove], slice2[indexToRemove+1:]...)
	return &slice2
}

func main() {
	ln, err := net.Listen(conn_type, host+":"+port)
	if err != nil {
		fmt.Println("Error when listening: ", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Listening on " + host + ":" + port)

	clients := []net.Conn{}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error while accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("New client connected:", conn.RemoteAddr())
		clients = append(clients, conn)
		fmt.Println(clients)
		go handleRequest(conn, &clients)
		fmt.Println(clients, "ok")
	}
}

func handleRequest(conn net.Conn, clients *[]net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client", conn.RemoteAddr(), "disconnected")
			conn.Close()
			clients = RemoveConn(clients, conn)
			break
		}
		fmt.Print(message)
		conn.Write([]byte("Message received: " + message))
	}
}
