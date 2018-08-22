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
