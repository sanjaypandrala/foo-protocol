package main

import (
	"bufio"
	"fmt"
	"net"
)

type Proxer interface {
	Execute(address string) error
}

type Proxy struct{}
var loopTimes = -1 // For testing purposes only.

func (m *Proxy) Execute(listenAddress, forwardAddress string) error {
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	defer ln.Close()
	i := 0
	for {
		if loopTimes > 0 {
			if i >= loopTimes {
				break
			}
			i++
		}
		connClient, _ := ln.Accept()
		go m.Communicate(connClient, forwardAddress, make(chan error))
	}
	return nil
}

func (m *Proxy) Communicate(connClient net.Conn, serverAddress string, c chan error) {
	connServer, err := net.Dial("tcp", serverAddress)
	if err != nil {
		c <- err
		return
	}
	defer connClient.Close()
	defer connServer.Close()
	serverReader := bufio.NewReader(connServer)
	clientReader := bufio.NewReader(connClient)
	for {
		msgClient, _ := clientReader.ReadBytes('\n')
		cType, err := m.validateMsg(msgClient)
		if err != nil {
			c <- err
			return
		}
		MetricsInstance.RecordMessage(cType)
		connServer.Write(msgClient)
		msgServer, _ := serverReader.ReadBytes('\n')
		sType, err := m.validateMsg(msgServer)
		if err != nil {
			c <- err
			return
		}
		MetricsInstance.RecordMessage(sType)
		connClient.Write(msgServer)
	}
	c <- nil
}

func (m *Proxy) validateMsg(msg []byte) (string, error) {
	if len(msg) <= 3 {
		return "", fmt.Errorf("Message lenght is incorrect")
	}
	msgType := string(msg)[:3]
	if msgType != "REQ" && msgType != "ACK" && msgType != "NAK" {
		return "", fmt.Errorf("Type must be REQ, ACK, or NAK")
	}
	// Not validating anything else since we need only the type
	return msgType, nil
}
