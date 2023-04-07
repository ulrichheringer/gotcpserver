package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	conn_type = "tcp"
	host      = "0.0.0.0"
	port      = "8080"
)

// Removes Connection from Array
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
		fmt.Println("Error when trying to start listening: ", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Listening on " + host + ":" + port)

	// Clients Array, contains: Ip, and conn reference
	clients := []net.Conn{}
	// Registered names from clients
	names := make(map[string]string)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error while accepting: ", err.Error())
			os.Exit(1)
		}
		timeLog := time.Now().Format(time.RFC822)
		fmt.Println(timeLog, "New client connected:", conn.RemoteAddr())
		// Adds client to the array
		clients = append(clients, conn)
		// Receives commands
		go handleRequest(conn, &clients, &names)
	}
}

func handleRequest(conn net.Conn, clients *[]net.Conn, names *map[string]string) {
main:
	for {
		// Reads incoming messages
		reader := bufio.NewReader(conn)
		message, err := reader.ReadString('\n')
		if err != nil {
			// If error, closes connection and removes client from array
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
		// Removes all \n, \r, \t, \b from the message for better support
		rNL := strings.ReplaceAll(message, "\n", "")
		rT := strings.ReplaceAll(rNL, "\t", "")
		rR := strings.ReplaceAll(rT, "\r", "")
		fUnquoted := strings.ReplaceAll(rR, "\b", "")
		// ==========================
		// Send command, syntax: /send (name of whom you wanna send a message) (message)
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
			// Register a name for the client
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
					conn.Write([]byte("You were sucessfully registered!\n"))
				}
			}
			// Command to show all registered users
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
