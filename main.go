package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	conn_type = "tcp"
	host      = "0.0.0.0"
	port      = "8080"
)

func RemoveConn(slice []net.Conn, conn net.Conn) []net.Conn {
	var indexToRemove int
	for i, s := range slice {
		if s == conn {
			indexToRemove = i
		}
	}
	return append(slice[:indexToRemove], slice[indexToRemove+1:]...)
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
		go handleRequest(conn, &clients)
	}
}

func handleRequest(conn net.Conn, clients *[]net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		message, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			fmt.Println("Client", conn.RemoteAddr(), "disconnected")
			*clients = RemoveConn(*clients, conn)
			break
		}
		// == Treating the message ==
		rNL := strings.ReplaceAll(message, "\n", "")
		rT := strings.ReplaceAll(rNL, "\t", "")
		rR := strings.ReplaceAll(rT, "\r", "")
		fUnquoted := strings.ReplaceAll(rR, "\b", "")
		// ==========================
		if fUnquoted == "show" {
			clientsLength := strconv.Itoa(len(*clients))
			clientsStr := fmt.Sprintf("%d\n", clients)
			conn.Write([]byte("Found " + clientsLength + " clients: " + clientsStr))
		} else if strings.HasPrefix(fUnquoted, "send") {
			if len(*clients) > 1 {
				send := strings.Fields(message)
				if len(send) > 2 {
					addr := send[1]
					for _, c := range *clients {
						if c.RemoteAddr().String() == addr {
							message := strings.Join(send[2:], " ")
							c.Write([]byte(conn.RemoteAddr().String() + " sent you a message: " + message + "\n"))
							break
						}
					}
					conn.Write([]byte("Client not found!\n"))
				} else {
					conn.Write([]byte("Wrong format! Please use: send 'addr(ip:port)'. Example: send 127.0.0.1:56120\n"))
				}
			} else {
				fmt.Println("not enough clients")
			}
		} else {
			conn.Write([]byte("Command not found.\n"))
		}
	}
}
