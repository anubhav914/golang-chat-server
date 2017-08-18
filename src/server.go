/*
things to do

create a struct omessage


*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatal(err)
	}
	msgchan := make(chan Message)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	// go printMessages(msgchan)
	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, msgchan, addchan, rmchan)
	}
}

func handleMessages(msgchan <-chan Message, addchan <-chan Client, rmchan <-chan Client) {
	chatrooms := make(map[string]map[net.Conn]chan<- string)

	for {
		select {
		case message := <-msgchan:
			for _, ch := range chatrooms[message.Group] {
				go func(mch chan<- string) { mch <- message.Msg }(ch)
			}
		case client := <-addchan:
			if _, ok := chatrooms[client.Group]; ok {
				chatrooms[client.Group][client.Conn] = client.Ch
			} else {
				chatrooms[client.Group] = map[net.Conn]chan<- string{client.Conn: client.Ch}
			}
		case expired_client := <-rmchan:
			delete(chatrooms[expired_client.Group], expired_client.Conn)
		}
	}

}

func handleConnection(c net.Conn, msgchan chan<- Message, addchan chan<- Client, rmchan chan<- Client) {
	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		Conn:     c,
		Nickname: promptNick(c, bufc),
		Group:    promptGroup(c, bufc),
		Ch:       make(chan string),
	}
	if strings.TrimSpace(client.Nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}
	if strings.TrimSpace(client.Group) == "" {
		io.WriteString(c, "Invalid GroupName\n")
		return
	}
	addchan <- client
	defer func() {
		msgchan <- Message{
			Msg:   fmt.Sprintf("User %s left the chat room\n", client.Nickname),
			Group: client.Group,
		}
		log.Printf("Connection from %v closed.\n", c.RemoteAddr())
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.Nickname))
	msgchan <- Message{
		Msg:   fmt.Sprintf("New user %s has joined the chat room\n", client.Nickname),
		Group: client.Group,
	}

	go client.ReadLinesInto(msgchan)
	client.WriteLinesFrom(client.Ch)
}

func promptNick(c net.Conn, bufc *bufio.Reader) string {
	io.WriteString(c, "Welcome to the fancy demo chat!\n")
	io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

func promptGroup(c net.Conn, bufc *bufio.Reader) string {
	io.WriteString(c, "What is your group? ")
	group, _, _ := bufc.ReadLine()
	return string(group)
}
