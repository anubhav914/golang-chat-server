package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

type Client struct {
	Conn     net.Conn
	Ch       chan string
	Nickname string
	Group    string
}

func (c Client) ReadLinesInto(ch chan<- Message) {
	bufc := bufio.NewReader(c.Conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}
		ch <- Message{
			Msg:   fmt.Sprintf("%s: %s", c.Nickname, line),
			Group: c.Group,
		}
	}
}

func (c Client) WriteLinesFrom(ch <-chan string) {
	for msg := range ch {
		_, err := io.WriteString(c.Conn, msg)
		if err != nil {
			return
		}
	}
}
