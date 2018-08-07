package main

import (
	"bytes"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/gopacket/tcpassembly"
	"os"
	"time"
)

const (
	SENDING   = 0
	RECEIVING = 1
)

type TcpStream struct {
	conversation         *TcpConversation
	start, end, complete bool
}

func (s *TcpStream) Reassembled(rs []tcpassembly.Reassembly) {
	s.conversation.Reassembled(s, rs)
}

func (s *TcpStream) ReassemblyComplete() {
	s.complete = true
	s.conversation.ReassemblyComplete(s)
}

type message struct {
	start, end int
	direction  uint8
	seen       time.Time
}

type TcpConversation struct {
	ipFlow               gopacket.Flow
	tcpFlow              gopacket.Flow
	srcStream, dstStream *TcpStream
	onComplete           func(c *TcpConversation)
	complete             bytes.Buffer
	messages             []message
	messageStart         int
	messageFirstSeen     time.Time
	lastSpeaker          *TcpStream
}

func (c *TcpConversation) Reassembled(s *TcpStream, rs []tcpassembly.Reassembly) {
	if c.lastSpeaker == nil {
		c.lastSpeaker = s
		c.messageFirstSeen = rs[0].Seen
	}
	if c.messageStart != c.complete.Len() && c.lastSpeaker != s {
		c.appendMessage()
		c.messageStart = c.complete.Len()
		c.messageFirstSeen = rs[0].Seen
		c.lastSpeaker = s
	}
	for _, r := range rs {
		s.start = s.start || r.Start
		s.end = s.end || r.End
		c.complete.Write(r.Bytes)
	}
}

func (c *TcpConversation)appendMessage() {
	var direction uint8
	if c.lastSpeaker == c.srcStream {
		direction = SENDING
	} else {
		direction = RECEIVING
	}
	c.messages = append(c.messages,
		message{c.messageStart, c.complete.Len(), direction, c.messageFirstSeen},
	)
}

func (c *TcpConversation) ReassemblyComplete(s *TcpStream) {
	if c.messageStart != c.complete.Len() {
		c.appendMessage()
		c.messageStart = c.complete.Len()
	}
	if c.srcStream != nil && c.dstStream != nil && c.srcStream.complete && c.dstStream.complete {
		c.onComplete(c)
	}
}

type key struct {
	net, transport gopacket.Flow
}

type TcpStreamFactory struct {
	completeConversation chan *TcpConversation
	conversationMap      map[key]*TcpConversation
}

func NewTcpStreamFactory() *TcpStreamFactory {
	return &TcpStreamFactory{
		completeConversation: make(chan *TcpConversation),
		conversationMap:      make(map[key]*TcpConversation),
	}
}

func (f *TcpStreamFactory) AddConversation(conversation *TcpConversation) {
	f.completeConversation <- conversation
}

func (f *TcpStreamFactory) GetConversation() *TcpConversation {
	return <-f.completeConversation
}

func (f *TcpStreamFactory) Close() {
	close(f.completeConversation)
}

func (f *TcpStreamFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	stream := &TcpStream{}

	k := key{netFlow, tcpFlow}
	conversation := f.conversationMap[k]
	if conversation == nil {
		conversation = &TcpConversation{
			ipFlow:     netFlow,
			tcpFlow:    tcpFlow,
			srcStream:  stream,
			onComplete: f.AddConversation,
		}
		f.conversationMap[key{netFlow.Reverse(), tcpFlow.Reverse()}] = conversation
	} else {
		delete(f.conversationMap, k)
		conversation.dstStream = stream
	}
	stream.conversation = conversation

	return stream
}

func ParseTcp(source *gopacket.PacketSource, factory *TcpStreamFactory) error {
	streamPool := tcpassembly.NewStreamPool(factory)

	// Limit memory usage by auto-flushing connection state if we get over 100K
	// packets in memory, or over 1000 for a single stream.
	assembler := tcpassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesTotal = 100000
	assembler.MaxBufferedPagesPerConnection = 1000

	defer func() {
		assembler.FlushAll()
		factory.Close()
	}()

	for true {
		packet, err := source.NextPacket()
		if err != nil {
			return err
		}
		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
			continue
		}
		net := packet.NetworkLayer().NetworkFlow()
		tcp := packet.TransportLayer().(*layers.TCP)
		ts := packet.Metadata().Timestamp
		assembler.AssembleWithTimestamp(net, tcp, ts)
	}
	return nil
}

type PcapParser struct {
	factory *TcpStreamFactory
	pcapFh  *os.File
	source  *gopacket.PacketSource

	errorChan chan error
}

func LoadPcap(filename string) (PcapParser, error) {
	var err error
	parser := PcapParser{
		factory:   NewTcpStreamFactory(),
		errorChan: make(chan error),
	}
	parser.pcapFh, err = os.Open(filename)
	if err != nil {
		return parser, err
	}
	reader, err := pcapgo.NewReader(parser.pcapFh)
	if err != nil {
		parser.pcapFh.Close()
		return parser, err
	}
	parser.source = gopacket.NewPacketSource(reader, reader.LinkType())
	return parser, nil
}

func (p PcapParser) Run() {
	go func() {
		p.errorChan <- ParseTcp(p.source, p.factory)
	}()
}

func (p PcapParser) GetConversation() (*TcpConversation, error) {
	var err error
	stream := p.factory.GetConversation()
	if stream == nil {
		err = <-p.errorChan
	}
	return stream, err
}

func (p PcapParser) Close() error {
	return p.pcapFh.Close()
}

func main() {}
