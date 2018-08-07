package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

var C = &TcpConversation{}
var ipFlow, _ = gopacket.FlowFromEndpoints(
	layers.NewIPEndpoint(net.IPv4(1, 1, 1, 1)),
	layers.NewIPEndpoint(net.IPv4(2, 2, 2, 2)),
)
var tcpFlow, _ = gopacket.FlowFromEndpoints(
	layers.NewTCPPortEndpoint(layers.TCPPort(1)),
	layers.NewTCPPortEndpoint(layers.TCPPort(2)),
)

func (c *TcpConversation) Messages() [][]byte {
	messages := make([][]byte, len(c.messages))
	buffer := c.complete.Bytes()
	for i, s := range c.messages {
		messages[i] = buffer[s.start:s.end]
	}
	return messages
}

func Test_TcpStream_UpdateFlags(t *testing.T) {
	s := &TcpStream{conversation: C}
	s.Reassembled([]tcpassembly.Reassembly{
		{},
	})

	assert.Equal(t, false, s.start)
	assert.Equal(t, false, s.end)

	s.Reassembled([]tcpassembly.Reassembly{
		{Start: true},
		{End: true},
	})

	assert.Equal(t, true, s.start)
	assert.Equal(t, true, s.end)
}

type ReassemblyTester struct {
	conversation      *TcpConversation
	request, response *TcpStream
	t                 *testing.T
}

func NewReassemblyTester(t *testing.T) ReassemblyTester {
	conversation := &TcpConversation{ipFlow: ipFlow, tcpFlow: tcpFlow}

	return ReassemblyTester{
		conversation: conversation,
		request:      &TcpStream{conversation: conversation},
		response:     &TcpStream{conversation: conversation},
		t:            t,
	}
}

func (test *ReassemblyTester) Messages(messages ...string) {
	for _, m := range messages {
		parts := strings.Split(m, "|")
		reassemblys := make([]tcpassembly.Reassembly, len(parts))
		for i, p := range parts {
			reassemblys[i] = tcpassembly.Reassembly{Bytes: []byte(p)}
		}
		if strings.Compare(m, ">EOF") == 0 {
			test.request.ReassemblyComplete()
		} else if strings.Compare(m, "<EOF") == 0 {
			test.response.ReassemblyComplete()
		} else if strings.HasPrefix(m, "<") {
			test.response.Reassembled(reassemblys)
		} else {
			test.request.Reassembled(reassemblys)
		}
	}
}

func (test *ReassemblyTester) Expect(expected ...string) {
	actual := test.conversation.Messages()
	strActual := make([]string, len(actual))
	for i, m := range actual {
		strActual[i] = string(m)
	}
	assert.Equal(test.t, expected, strActual)
}

func Test_TcpStream_Reassemble_1(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello", "<EOF", ">EOF")
	test.Expect(">hello")
}

func Test_TcpStream_Reassemble_2(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello", ">EOF", "<EOF")
	test.Expect(">hello")
}

func Test_TcpStream_Reassemble_3(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages("<hello", "<EOF", ">EOF")
	test.Expect("<hello")
}

func Test_TcpStream_Reassemble_4(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages("<hello", ">EOF", "<EOF")
	test.Expect("<hello")
}

func Test_TcpStream_Reassemble_5(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello| friend", "<EOF", ">EOF")
	test.Expect(">hello friend")
}

func Test_TcpStream_Reassemble_5_1(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages("<hello", "<friend|how", "<are", "<you", "<EOF", ">EOF")
	test.Expect("<hello<friendhow<are<you")
}

func Test_TcpStream_Reassemble_6(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages("<hello|friend", "<EOF", ">EOF")
	test.Expect("<hellofriend")
}

func Test_TcpStream_Reassemble_7(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello", ">friend", "<EOF", ">EOF")
	test.Expect(">hello>friend")
}

func Test_TcpStream_Reassemble_8(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages("<hello", "<friend", "<EOF", ">EOF")
	test.Expect("<hello<friend")
}

func Test_TcpStream_Reassemble_9(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello", "<hello", ">friend", ">EOF", "<EOF")
	test.Expect(">hello", "<hello", ">friend")
}

func Test_TcpStream_Reassemble_10(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello", "<hello", ">friend", "<friend", ">EOF", "<EOF")
	test.Expect(">hello", "<hello", ">friend", "<friend")
}

func Test_TcpStream_Reassemble_11(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages(">hello", ">friend", "<hello", ">how", ">are", ">you", ">EOF", "<EOF")
	test.Expect(">hello>friend", "<hello", ">how>are>you")
}

func Test_TcpStream_NoEmptyMessages(t *testing.T) {
	test := NewReassemblyTester(t)
	test.Messages("", "<hello", ">EOF", "<EOF")
	test.Expect("<hello")
}

func Test_TcpStream_SetSeenOnFirstStream(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(1 * time.Second)

	conversation := &TcpConversation{ipFlow: ipFlow, tcpFlow: tcpFlow}
	request := &TcpStream{conversation: conversation}
	request.Reassembled([]tcpassembly.Reassembly{
		{Bytes: []byte("something"), Seen: t1},
	})
	response := &TcpStream{conversation: conversation}
	response.Reassembled([]tcpassembly.Reassembly{
		{Bytes: []byte("something"), Seen: t2},
	})
	response.ReassemblyComplete()
	request.ReassemblyComplete()

	assert.Equal(t, t1, conversation.messages[0].seen)
	assert.Equal(t, t2, conversation.messages[1].seen)
}

func Test_TcpStreamFactory_SetStreams(t *testing.T) {
	f := NewTcpStreamFactory()

	request := f.New(ipFlow, tcpFlow).(*TcpStream)

	assert.Equal(t, tcpFlow, request.conversation.tcpFlow)
	assert.Equal(t, ipFlow, request.conversation.ipFlow)
}

func Test_TcpStreamFactory_SetRequestResponse(t *testing.T) {
	f := NewTcpStreamFactory()

	request := f.New(ipFlow, tcpFlow).(*TcpStream)
	response := f.New(ipFlow.Reverse(), tcpFlow.Reverse()).(*TcpStream)

	assert.Equal(t, request.conversation, response.conversation)
	assert.Equal(t, request, response.conversation.srcStream)
	assert.Equal(t, response, request.conversation.dstStream)
}

func Test_Parser_Integration(t *testing.T) {
	done := sync.WaitGroup{}
	done.Add(1)
	f := NewTcpStreamFactory()
	defer f.Close()
	var conversation *TcpConversation
	go func() {
		conversation = f.GetConversation()
		done.Done()
	}()

	stream := f.New(ipFlow, tcpFlow)
	stream.ReassemblyComplete()
	streamReverse := f.New(ipFlow.Reverse(), tcpFlow.Reverse())
	streamReverse.ReassemblyComplete()
	done.Wait()

	assert.Equal(t, ipFlow, conversation.ipFlow)
	assert.Equal(t, tcpFlow, conversation.tcpFlow)
}

func Test_Parser_Acceptance_IPv4(t *testing.T) {
	parser, err := LoadPcap("../test/ipv4_tcp_simple.pcap")
	assert.Nil(t, err)
	defer parser.Close()

	parser.Run()

	c, err := parser.GetConversation()
	assert.Nil(t, err)
	assert.Equal(t, []byte(">hello\n<hello\n>friend\n<friend\n"), c.complete.Bytes())
	seen := time.Date(2018, 07, 15, 17, 13, 0, 0, time.UTC)
	assert.Equal(t, []message{
		{0, 7, SENDING, seen.Add(2611587 * time.Microsecond)},
		{7, 14, RECEIVING,seen.Add(10561471 * time.Microsecond)},
		{14, 22, SENDING,seen.Add(16259779 * time.Microsecond)},
		{22, 30, RECEIVING,seen.Add(20339612 * time.Microsecond)},
	}, c.messages)
	assert.True(t, c.srcStream.start && c.srcStream.end)
	assert.True(t, c.dstStream.start && c.dstStream.end)
	assert.Equal(t, "127.0.0.1->127.0.0.1", fmt.Sprint(c.ipFlow))
	assert.Equal(t, "42266->8080", fmt.Sprint(c.tcpFlow))

	c, err = parser.GetConversation()
	assert.Nil(t, c)
	assert.Equal(t, io.EOF, err)
}

func Test_Parser_Acceptance_IPv6(t *testing.T) {
	parser, err := LoadPcap("../test/ipv6_tcp_simple.pcap")
	assert.Nil(t, err)
	defer parser.Close()

	parser.Run()

	c, err := parser.GetConversation()
	assert.Nil(t, err)
	assert.Equal(t, []byte(">hi\n<hi\n>there\n<there\n"), c.complete.Bytes())
	seen := time.Date(2018, 07, 15, 17, 13, 0, 0, time.UTC)
	assert.Equal(t, []message{
		{0, 4, SENDING,seen.Add(30339421 * time.Microsecond)},
		{4, 8, RECEIVING,seen.Add(38753811 * time.Microsecond)},
		{8, 15, SENDING,seen.Add(43490711 * time.Microsecond)},
		{15, 22, RECEIVING,seen.Add(47410317 * time.Microsecond)},
	}, c.messages)
	assert.True(t, c.srcStream.start && c.srcStream.end)
	assert.True(t, c.dstStream.start && c.dstStream.end)
	assert.Equal(t, "::1->::1", fmt.Sprint(c.ipFlow))
	assert.Equal(t, "40980->8080", fmt.Sprint(c.tcpFlow))

	c, err = parser.GetConversation()
	assert.Nil(t, c)
	assert.Equal(t, io.EOF, err)
}
