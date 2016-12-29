package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
)

type ProxyTestSuite struct {
	suite.Suite
}

func (s *ProxyTestSuite) SetupTest() {}

type clientArgs struct {
	msgIn  string
	msgOut string
}
type serverArgs struct {
	address string
}

var clientChannel = make(chan clientArgs)
var serverChannel = make(chan serverArgs)

func TestProxyUnitTestSuite(t *testing.T) {
	s := new(ProxyTestSuite)
	defer func() {
		clientChannel <- clientArgs{}
		serverChannel <- serverArgs{}
	}()
	// Client mock
	go func() {
		for {
			args := <-clientChannel
			conn, err := net.Dial("tcp", "localhost:8002")
			if err != nil {
				break
			}
			if _, err := conn.Write([]byte(args.msgIn)); err != nil {
				break
			}
			reader := bufio.NewReader(conn)
			msgOut, _ := reader.ReadBytes('\n')
			args.msgOut = string(msgOut)
			clientChannel <- args
			conn.Close()
		}
	}()
	// Server mock
	go func() {
		for {
			args := <-serverChannel
			ln, err := net.Listen("tcp", args.address)
			if err != nil {
				return
			}
			conn, err := ln.Accept()
			reader := bufio.NewReader(conn)
			msg, _ := reader.ReadBytes('\n')
			retMsg := bytes.Replace(msg, []byte{'R', 'E', 'Q'}, []byte{'A', 'C', 'K'}, -1)
			conn.Write(retMsg)
			ln.Close()
		}
	}()

	suite.Run(t, s)
}

// Communicate

func (s *ProxyTestSuite) Test_Communicate_ForwardsMessages() {
	clientChannel <- clientArgs{msgIn: "REQ 1 Foo\n"}
	serverChannel <- serverArgs{address: ":8001"}
	expected := "ACK 1 Foo\n"
	p := Proxy{}
	ln, _ := net.Listen("tcp", ":8002")
	conn, _ := ln.Accept()
	defer ln.Close()

	go p.Communicate(conn, "localhost:8001", make(chan error))
	actual := <-clientChannel

	s.Equal(expected, actual.msgOut)
}

func (s *ProxyTestSuite) Test_Communicate_ReturnsError_WhenServerIsNotAvailable() {
	p := Proxy{}
	actual := make(chan error)

	go p.Communicate(nil, "localhost:8001", actual)

	s.Error(<-actual)
}

func (s *ProxyTestSuite) Test_Communicate_ReturnsError_WhenFormatIsIncorrect() {
	testData := []struct {
		msg string
	}{
		{msg: "x\n"}, // Too short
		{msg: "FOO 1 Is not a valid type\n"},
	}
	for _, t := range testData {
		actual := make(chan error)
		clientChannel <- clientArgs{msgIn: t.msg}
		serverChannel <- serverArgs{address: ":8001"}
		p := Proxy{}
		ln, _ := net.Listen("tcp", ":8002")
		conn, _ := ln.Accept()

		go p.Communicate(conn, "localhost:8001", actual)
		s.Error(<-actual)
		<-clientChannel
		ln.Close()
	}
}

func (s *ProxyTestSuite) Test_Communicate_RecordsMessages() {
	miOrig := MetricsInstance
	defer func() { MetricsInstance = miOrig }()
	actual := []string{}
	miMock := MetricsMock{
		RecordMessageMock: func(msgType string) {
			actual = append(actual, msgType)
		},
	}
	MetricsInstance = miMock
	clientChannel <- clientArgs{
		msgIn: "REQ 1 Foo\n",
	}
	serverChannel <- serverArgs{
		address: ":8001",
	}
	p := Proxy{}
	ln, _ := net.Listen("tcp", ":8002")
	conn, _ := ln.Accept()
	defer ln.Close()

	go p.Communicate(conn, "localhost:8001", make(chan error))
	<-clientChannel

	s.Len(actual, 2)
	s.Contains(actual, "REQ")
	s.Contains(actual, "ACK")
}

// Execute

func (s *ProxyTestSuite) Test_Execute_RunsCommunicateAsALoop() {
	p := Proxy{}
	loopTimes = 20

	go p.Execute(":8002", "localhost:8001")

	for i := 0; i < loopTimes; i++ {
		clientChannel <- clientArgs{msgIn: fmt.Sprintf("REQ %d Foo\n", i)}
		serverChannel <- serverArgs{address: ":8001"}
		expected := fmt.Sprintf("ACK %d Foo\n", i)

		actual := <-clientChannel

		s.Equal(expected, actual.msgOut)
	}
}

func (s *ProxyTestSuite) Test_xExecute_ReturnsError_WhenListeningFails() {
	ln, _ := net.Listen("tcp", ":8002")
	defer ln.Close()
	p := Proxy{}

	actual := p.Execute(":8002", "localhost:8001")

	s.Error(actual)
}
