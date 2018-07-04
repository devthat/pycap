import ipaddress
import os
import unittest
from datetime import datetime, timedelta

from pycap import Message, Pcap


def locate(x):
    return os.path.join(os.path.dirname(os.path.realpath(__file__)), x)


class MessageTest(unittest.TestCase):
    def test_repr(self):
        now = datetime(year=2017, month=11, day=13)

        m = Message(now, "sending", b"message")

        self.assertEqual("Message(2017-11-13 00:00:00,sending,b'message')", repr(m))

    def test_eq(self):
        now = datetime(year=2017, month=11, day=13)
        m = Message(now, "sending", b"message")

        self.assertEqual(m, m)
        self.assertNotEqual(Message(now + timedelta(seconds=1), "sending", b"message"), m)
        self.assertNotEqual(Message(now, "receiving", b"message"), m)
        self.assertNotEqual(Message(now, "sending", b"different"), m)
        self.assertNotEqual(m, (now, "sending", b"message"))

    def test_hash(self):
        now = datetime(year=2017, month=11, day=13)
        m1 = Message(now, "sending", b"message")
        m2 = Message(now, "sending", b"message")

        self.assertEqual(hash(m1), hash(m2))
        self.assertEqual({m1: "a"}, {m2: "a"})


class AcceptanceTest(unittest.TestCase):
    def test_ipv4(self):
        with Pcap(locate("ipv4_tcp_simple.pcap")) as p:
            conversation, = list(p)

            self.assertEqual(ipaddress.IPv4Address("127.0.0.1"), conversation.srcIp)
            self.assertEqual(42266, conversation.srcPort)
            self.assertEqual(ipaddress.IPv4Address("127.0.0.1"), conversation.dstIp)
            self.assertEqual(8080, conversation.dstPort)
            self.assertEqual(b">hello\n<hello\n>friend\n<friend\n", conversation.complete)
            self.assertEqual([
                Message(datetime(2018, 7, 15, 19, 13, 2), "sending", b'>hello\n'),
                Message(datetime(2018, 7, 15, 19, 13, 10), "receiving", b'<hello\n'),
                Message(datetime(2018, 7, 15, 19, 13, 16), "sending", b'>friend\n'),
                Message(datetime(2018, 7, 15, 19, 13, 20), "receiving", b'<friend\n'),
            ], conversation.dialog)


if __name__ == "__main__":
    unittest.main()
