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

typedef unsigned char GoUint8;
extern GoUint8 pcap_load(void** p0, char* p1);
extern GoUint8 pcap_close(void* p0);
extern char* pcap_error(void* p0);
extern GoUint8 pcap_parse(void* p0, struct TcpConversation* p1);
