package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

/*
#include <stdlib.h>
*/
import "C"

func cChar(goString string) *C.char {
	return C.CString(goString)
}

func goString(cString *C.char) string {
	return C.GoString(cString)
}

func goBytes(cString *C.uchar, len int) []byte {
	return C.GoBytes(unsafe.Pointer(cString), C.int(len))
}

func cFree(p unsafe.Pointer) {
	C.free(p)
}

func assertNil(t *testing.T, p *C.char) {
	assert.Equal(t, (*C.char)(nil), p)
}
