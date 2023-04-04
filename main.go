package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
	names := make(map[string]string)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error while accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("New client connected:", conn.RemoteAddr())
		clients = append(clients, conn)
		go handleRequest(conn, &clients, &names)
	}
}

func handleRequest(conn net.Conn, clients *[]net.Conn, names *map[string]string) {
main:
	for {
		reader := bufio.NewReader(conn)
		message, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			fmt.Println("Client", conn.RemoteAddr(), "disconnected")
			*clients = RemoveConn(*clients, conn)
			for user, ip := range *names {
				if ip == conn.RemoteAddr().String() {
					delete(*names, user)
				}
			}
			break
		}
		// == Treating the message ==
		rNL := strings.ReplaceAll(message, "\n", "")
		rT := strings.ReplaceAll(rNL, "\t", "")
		rR := strings.ReplaceAll(rT, "\r", "")
		fUnquoted := strings.ReplaceAll(rR, "\b", "")
		// ==========================
		if strings.HasPrefix(fUnquoted, "/send") {
			if len(*clients) > 1 {
				send := strings.Fields(message)
				if len(send) > 2 {
					receiver := send[1]
					var sender string
					for senderName, senderIp := range *names {
						if conn.RemoteAddr().String() == senderIp {
							sender = senderName
							break
						}
					}
					if len(sender) == 0 {
						conn.Write([]byte("You need to be registered to send a message\n"))
						continue main
					}
					for name, ip := range *names {
						if name == receiver {
							message := strings.Join(send[2:], " ")
							for _, c := range *clients {
								if c.RemoteAddr().String() == ip {
									c.Write([]byte(sender + " sent you a message: " + message + "\n"))
									continue main
								}
							}
						}
					}
					conn.Write([]byte("Client not found!\n"))
				} else {
					conn.Write([]byte("Wrong format! Use: /send 'addr(ip:port)' message.\n"))
				}
			} else {
				fmt.Println("not enough clients")
			}
		} else if strings.HasPrefix(fUnquoted, "/register") {
			if len(strings.Fields(fUnquoted)) > 1 {
				name := strings.Fields(fUnquoted)[1]
				if _, ok := (*names)[name]; ok {
					conn.Write([]byte("The username " + name + " is already in use\n"))
				} else {
					userIp := conn.RemoteAddr().String()
					for username, ip := range *names {
						if ip == userIp {
							conn.Write([]byte("You are already registered with the " + username + " username.\n"))
							continue main
						}
					}
					(*names)[name] = userIp
					conn.Write([]byte("Você foi registrado com sucesso!\n"))
				}
			}
		} else if fUnquoted == "/users" {
			var users []string
			for key := range *names {
				users = append(users, key)
			}
			if len(users) < 1 {
				conn.Write([]byte("There aren't any users online\n"))
				continue main
			}
			usersStringify := strings.Join(users, ",")
			message := fmt.Sprintf("There are %d users online: %s\n", len(users), usersStringify)
			conn.Write([]byte(message))
		} else {
			conn.Write([]byte("Command not found.\n"))
		}
	}
}
