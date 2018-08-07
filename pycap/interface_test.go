package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"unsafe"
)

func Test_Interface_ErrorIsReturned(t *testing.T) {
	var p unsafe.Pointer

	result := pcap_load(&p, cChar("not a file"))
	assert.Equal(t, failed, result)

	err := pcap_error(p)
	assert.True(t, strings.HasPrefix(goString(err), "open"))
	cFree(unsafe.Pointer(err))
}

func Test_Interface_ParseFile_IPv4(t *testing.T) {
	var p unsafe.Pointer
	c := tcpConversation()

	result := pcap_load(&p, cChar("../test/ipv4_tcp_simple.pcap"))
	assert.Equal(t, success, result)

	result = pcap_parse(p, &c)
	assert.Equal(t, success, result)

	assert.Equal(t, uint32(0x7f000001), uint32(c.src.ip32))
	assert.Equal(t, uint32(42266), uint32(c.src.port))
	assert.Equal(t, uint32(0x7f000001), uint32(c.dst.ip32))
	assert.Equal(t, uint32(8080), uint32(c.dst.port))
	assert.Equal(t, uint8(0xcc), uint8(c.flags))
	assert.Equal(t, []byte(">hello\n<hello\n>friend\n<friend\n"), goBytes(c.complete, 30))
	assert.Equal(t, uint32(30), uint32(c.len))
	assert.Equal(t,
		[]GoMessage{
			{0 , 7 , 1531674782 ,0, [7]byte{}},
			{7 , 14,  1531674790, 1, [7]byte{}},
			{14 , 22,  1531674796, 0, [7]byte{}},
			{22 , 30,  1531674800, 1, [7]byte{}},
		},
		asMessages(c.messages)[:4])
	assert.Equal(t, uint32(4), uint32(c.num))

	result = pcap_parse(p, &c)
	assert.Equal(t, failed, result)

	err := pcap_error(p)
	assertNil(t, err)

	result = pcap_close(p)
	assert.Equal(t, success, result)

	cFree(unsafe.Pointer(c.messages))
	cFree(unsafe.Pointer(c.complete))
}

func Test_Interface_ParseFile_IPv6(t *testing.T) {
	var p unsafe.Pointer
	c := tcpConversation()

	result := pcap_load(&p, cChar("../test/ipv6_tcp_simple.pcap"))
	assert.Equal(t, success, result)

	result = pcap_parse(p, &c)
	assert.Equal(t, success, result)

	assert.Equal(t, uint32(0x0), uint32(c.src.ip32))
	assert.Equal(t, uint32(0x0), uint32(c.src.ip64))
	assert.Equal(t, uint32(0x0), uint32(c.src.ip96))
	assert.Equal(t, uint32(0x1), uint32(c.src.ip128))
	assert.Equal(t, uint32(40980), uint32(c.src.port))
	assert.Equal(t, uint32(0x0), uint32(c.dst.ip32))
	assert.Equal(t, uint32(0x0), uint32(c.dst.ip64))
	assert.Equal(t, uint32(0x0), uint32(c.dst.ip96))
	assert.Equal(t, uint32(0x1), uint32(c.dst.ip128))
	assert.Equal(t, uint32(8080), uint32(c.dst.port))
	assert.Equal(t, uint8(0xcc), uint8(c.flags))
	assert.Equal(t, []byte(">hi\n<hi\n>there\n<there\n"), goBytes(c.complete, 22))
	assert.Equal(t, uint32(22), uint32(c.len))
	assert.Equal(t,
		[]GoMessage{
			{0, 4, 1531674810, 0, [7]byte{}},
			{4, 8, 1531674818, 1, [7]byte{}},
			{8, 15, 1531674823, 0, [7]byte{}},
			{15, 22, 1531674827, 1, [7]byte{}},
	}, asMessages(c.messages)[:4])
	assert.Equal(t, uint32(4), uint32(c.num))

	result = pcap_parse(p, &c)
	assert.Equal(t, failed, result)

	err := pcap_error(p)
	assertNil(t, err)

	result = pcap_close(p)
	assert.Equal(t, success, result)

	cFree(unsafe.Pointer(c.messages))
	cFree(unsafe.Pointer(c.complete))
}
