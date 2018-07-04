cdef extern from "libgocap.h":
    cdef struct Endpoint:
        unsigned int len;
        unsigned int ip32;
        unsigned int ip64;
        unsigned int ip96;
        unsigned int ip128;
        unsigned int port;

    cdef struct Message:
        unsigned int start;
        unsigned int end;
        unsigned long ts;
        unsigned char direction;

    cdef struct TcpConversation:
        Endpoint src;
        Endpoint dst;
        unsigned char flags;
        unsigned char *complete;
        unsigned int len;
        Message *messages;
        unsigned int num;

    unsigned char pcap_load(void** p0, char* p1);
    unsigned char pcap_parse(void* p0, TcpConversation* p1);
    unsigned char pcap_close(void* p0);
    char* pcap_error(void* p0);
