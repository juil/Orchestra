/* marshal.go
 *
 * Common wire marshalling code.
*/

package orchestra;

import (
	"os"
	"net"
	"goprotobuf.googlecode.com/hg/proto"
)

type WirePkt struct {
	Type	byte
	Length	uint16
	Payload []byte
}

const (
	TypeNop = 0
	TypeIdentifyClient = 1
	TypeReadyForTask = 2
	TypeTaskRequest = 3
)

var (
	ErrMalformedMessage = os.NewError("Malformed Message")
	ErrUnknownMessage   = os.NewError("Unknown Message")
)

func (p *WirePkt) Send(c *net.Conn) (n int, err os.Error) {
	n = 0
	preamble := make([]byte, 3)
	preamble[0] = p.Type
	preamble[1] = byte((p.Length >> 8) & 0xFF)
	preamble[2] = byte(p.Length & 0xFF)
	ninc, err := (*c).Write(preamble)
	n += ninc
	if (err != nil) {
		return n, err
	}	
	ninc, err = (*c).Write(p.Payload[0:p.Length])
	n += ninc
	if (err != nil) {
		return n, err
	}
	return n, nil
}

func (p *WirePkt) Decode() (obj interface{}, err os.Error) {
	switch (p.Type) {
	case TypeNop:
		if (p.Length != 0) {
			/* throw error later... */
			return nil, ErrMalformedMessage;
		}
	case TypeIdentifyClient:
		ic := new(IdentifyClient)
		proto.Unmarshal(p.Payload[0:p.Length], ic)
		return ic, nil
	}
	return nil, ErrUnknownMessage
}

func MakeNop() (p *WirePkt) {
	p = new(WirePkt)
	p.Length = 0
	p.Type = TypeNop
	p.Payload = nil

	return p
}

func MakeIdentifyClient(hostname string) (p *WirePkt) {
	p = new(WirePkt)
	s := new(IdentifyClient)
	s.Hostname = new(string)
	*(s.Hostname) = hostname
	
	var err os.Error
	p.Payload, err = proto.Marshal(s)
	if (err != nil) {
		return nil
	}
	if len(p.Payload) >= 0x10000 {
		/* result is too big to encode */
		return nil
	}
	p.Length = uint16(len(p.Payload))

	return p
}

func Receive(c *net.Conn) (msg *WirePkt, err os.Error) {
	msg = new(WirePkt)
	preamble := make([]byte, 3)

	n, err := (*c).Read(preamble)
	if (n < 3) {
		/* short read!  wtf! err? */
		return nil, err
	}
	msg.Type = preamble[0]
	msg.Length = (uint16(preamble[1]) << 8) | uint16(preamble[2])
	msg.Payload = make([]byte, msg.Length)
	n, err = (*c).Read(msg.Payload)

	/* Decode! */
	return msg, nil
}


