package main

// TODO: what about Conversation containing \x0?
// TODO: nested protocols
// TODO: errors need to be freed

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"io"
	"unsafe"
)

/*
#include <stdlib.h>

struct Endpoint {
	unsigned int len;
	unsigned int ip32;
	unsigned int ip64;
	unsigned int ip96;
	unsigned int ip128;
	unsigned int port;
};

struct Message {
	unsigned int start;
	unsigned int end;
	unsigned long ts;
	unsigned char direction;
};

struct TcpConversation {
	struct Endpoint src;
	struct Endpoint dst;
	unsigned char flags;
	unsigned char *complete;
	unsigned int len;
	struct Message *messages;
	unsigned int num;
};
*/
import "C"

const (
	success = uint8(0)
	failed  = uint8(1)
)

type GoMessage C.struct_Message

type PcapParserInterface struct {
	PcapParser
	err error
}

//export pcap_load
func pcap_load(parser *unsafe.Pointer, filename *C.char) uint8 {
	pi := PcapParserInterface{}
	*parser = unsafe.Pointer(&pi)

	pi.PcapParser, pi.err = LoadPcap(C.GoString(filename))
	if pi.err != nil {
		return failed
	}

	pi.Run()
	return success
}

//export pcap_close
func pcap_close(parser unsafe.Pointer) uint8 {
	p := (*PcapParserInterface)(parser)
	p.err = p.Close()
	if p.err != nil {
		return failed
	}
	return success
}

//export pcap_error
func pcap_error(parser unsafe.Pointer) *C.char {
	p := (*PcapParserInterface)(parser)
	if p.err == nil || p.err == io.EOF {
		return nil
	}
	return C.CString(p.err.Error())
}

//export pcap_parse
func pcap_parse(parser unsafe.Pointer, conversation *C.struct_TcpConversation) uint8 {
	p := (*PcapParserInterface)(parser)
	var gonversation *TcpConversation
	gonversation, p.err = p.GetConversation()
	if gonversation == nil {
		return failed
	}
	setIp(&conversation.src, gonversation.ipFlow.Src())
	conversation.src.port = C.uint(binary.BigEndian.Uint16(gonversation.tcpFlow.Src().Raw()))
	setIp(&conversation.dst, gonversation.ipFlow.Dst())
	conversation.dst.port = C.uint(binary.BigEndian.Uint16(gonversation.tcpFlow.Dst().Raw()))
	conversation.flags = 0x0
	if gonversation.srcStream.start {
		conversation.flags |= 0x80
	}
	if gonversation.srcStream.end {
		conversation.flags |= 0x40
	}
	if gonversation.dstStream.start {
		conversation.flags |= 0x08
	}
	if gonversation.dstStream.end {
		conversation.flags |= 0x04
	}
	complete := gonversation.complete.Bytes()
	conversation.complete = (*C.uchar)(C.CBytes(complete))
	conversation.len = C.uint(len(complete))

	size := C.sizeof_struct_Message * len(gonversation.messages)
	conversation.messages = (*C.struct_Message)(C.malloc(C.ulong(size)))
	messages := asMessages(conversation.messages)
	for i, m := range gonversation.messages {
		messages[i].start = C.uint(m.start)
		messages[i].end = C.uint(m.end)
		messages[i].ts = C.ulong(m.seen.Unix())
		messages[i].direction = C.uchar(m.direction)
	}
	conversation.num = C.uint(len(gonversation.messages))
	return success
}

func setIp(endpoint *C.struct_Endpoint, ipEnd gopacket.Endpoint) {
	ip := ipEnd.Raw()
	switch ipEnd.EndpointType() {
	case layers.EndpointIPv4:
		endpoint.len = 4
		endpoint.ip32 = C.uint(binary.BigEndian.Uint32(ip[0:4]))
	case layers.EndpointIPv6:
		endpoint.len = 16
		endpoint.ip32 = C.uint(binary.BigEndian.Uint32(ip[0:4]))
		endpoint.ip64 = C.uint(binary.BigEndian.Uint32(ip[4:8]))
		endpoint.ip96 = C.uint(binary.BigEndian.Uint32(ip[8:12]))
		endpoint.ip128 = C.uint(binary.BigEndian.Uint32(ip[12:16]))
	}
}

func tcpConversation() C.struct_TcpConversation {
	var conversation C.struct_TcpConversation
	return conversation
}

func asMessages(p *C.struct_Message) *[1 << 30]GoMessage {
	return (*[1 << 30]GoMessage)(unsafe.Pointer(p))
}
