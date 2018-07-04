# distutils: library_dirs = .
# distutils: include_dirs = .

import ipaddress

from cpython.datetime cimport datetime, PyDateTime_CAPI
from Cython.Shadow import boundscheck, wraparound
from libc.stdlib cimport malloc, free
cimport cpycap

cdef class Message:
    cdef readonly datetime time
    cdef readonly str direction
    cdef readonly bytes message

    def __cinit__(self, datetime time, str direction, bytes message):
        self.time = time
        self.direction = direction
        self.message = message

    def __repr__(self):
        return "{}({},{},{})".format(self.__class__.__name__, self.time, self.direction, self.message)

    def __eq__(self, other):
        if isinstance(other, self.__class__):
            return self.time == other.time and self.direction == other.direction and self.message == other.message
        return False

    def __hash__(self):
        return hash((self.time, self.direction, self.message))


cdef class TcpConversation:
    cdef cpycap.TcpConversation *_ptr
    cdef list _dialog

    def __init__(self):
        self._dialog = None

    def __dealloc__(self):
        free(self._ptr.complete)
        free(self._ptr.messages)
        free(self._ptr)

    @staticmethod
    cdef TcpConversation from_ptr(cpycap.TcpConversation *ptr):
        cdef TcpConversation c = TcpConversation.__new__(TcpConversation)
        c._ptr = ptr
        return c

    @property
    def srcIp(self):
        return ipaddress.ip_address(self._ptr.src.ip32)

    @property
    def srcPort(self):
        return self._ptr.src.port

    @property
    def dstIp(self):
        return ipaddress.ip_address(self._ptr.dst.ip32)

    @property
    def dstPort(self):
        return self._ptr.dst.port

    @property
    def complete(self):
        return self._ptr.complete[:self._ptr.len]

    @property
    def dialog(self):
        if self._dialog is None:
            self._dialog = self._parse_dialog()
        return self._dialog

    cdef list _parse_dialog(self):
        dialog = []
        with boundscheck(False), wraparound(False):
            for i in range(self._ptr.num):
                m = self._ptr.messages[i]
                when = datetime.fromtimestamp(m.ts)
                message_text = self._ptr.complete[m.start:m.end]
                direction = "receiving" if m.direction else "sending"
                message = Message.__new__(Message, when, direction, message_text)
                dialog.append(message)
        return dialog


cdef class Pcap:
    cdef void *p

    def __cinit__(self, str filename):
        cpycap.pcap_load(&self.p, filename.encode())

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.close()

    def __iter__(self):
        return self

    def __next__(self):
        cdef char *err
        cdef cpycap.TcpConversation *c = <cpycap.TcpConversation *>malloc(sizeof(cpycap.TcpConversation))
        if cpycap.pcap_parse(self.p, c):
            err = cpycap.pcap_error(self.p)
            if err != NULL:
                raise RuntimeError(err.decode())
            raise StopIteration()
        return TcpConversation.from_ptr(c)

    cpdef close(self):
        cpycap.pcap_close(self.p)
