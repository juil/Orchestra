/* wire.go
 *
 * Wire Level Encapsulation
*/

package orchestra;

import (
	"os"
	"net"
	"fmt"
)

type WirePkt struct {
	Type	byte
	Length	uint16
	Payload []byte
}

const (
	TypeNop			= 0
	TypeIdentifyClient	= 1
	TypeReadyForTask	= 2
	TypeTaskRequest		= 3
	TypeTaskResponse	= 4
	TypeAcknowledgement	= 5
)

var (
	ErrMalformedMessage = os.NewError("Malformed Message")
	ErrUnknownMessage   = os.NewError("Unknown Message")
)

func (p *WirePkt) ValidUnidentified() bool {
	if p.Type == TypeNop {
		return true
	}
	if p.Type == TypeIdentifyClient {
		return true
	}

	return false
}

func (p *WirePkt) Send(c net.Conn) (n int, err os.Error) {
	n = 0
	preamble := make([]byte, 3)
	preamble[0] = p.Type
	preamble[1] = byte((p.Length >> 8) & 0xFF)
	preamble[2] = byte(p.Length & 0xFF)
	ninc, err := c.Write(preamble)
	n += ninc
	if (err != nil) {
		return n, err
	}	
	ninc, err = c.Write(p.Payload[0:p.Length])
	n += ninc
	if (err != nil) {
		return n, err
	}
	return n, nil
}

func (p *WirePkt) Dump() {
	fmt.Printf("Packet Dump: Type %d, Len %d\n", p.Type, p.Length)
	for i := 0; i < int(p.Length); i++ {
		if i%16 == 0 {
			fmt.Printf("%04x: ", i)
		}
		fmt.Printf("%02x ", p.Payload[i])
		if i%16 == 15 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func Receive(c net.Conn) (msg *WirePkt, err os.Error) {
	msg = new(WirePkt)
	preamble := make([]byte, 3)

	n, err := c.Read(preamble)
	if err != nil {
		return nil, err
	}
	if n < 3 {
		/* short read!  wtf! err? */
		return nil, ErrMalformedMessage
	}
	msg.Type = preamble[0]
	msg.Length = (uint16(preamble[1]) << 8) | uint16(preamble[2])
	if msg.Length > 0 {
		msg.Payload = make([]byte, msg.Length)
		n, err = c.Read(msg.Payload)
		if err != nil {
			return nil, err
		}
		if n < int(msg.Length) {
			/* short read!  wtf! err? */
			return nil, ErrMalformedMessage
		}
	}

	/* Decode! */
	return msg, nil
}

