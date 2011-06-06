/* messages.go
 *
 * Messages, Marshalling, etc
*/

package orchestra

import (
	"os"
	"goprotobuf.googlecode.com/hg/proto"
)

/* we define this message to make sending easy */
type Message interface {
	Marshal()	interface{}
	Type()		uint8
}

var ErrMessageTooLarge = os.NewError("Marshalled Message is too large on wire")

func Encode(msg Message) (pkt *WirePkt, err os.Error) {
	pkt = new(WirePkt)
	pkt.Type = msg.Type()

	mo := msg.Marshal()

	if mo != nil {
		var err os.Error
		pkt.Payload, err = proto.Marshal(mo)
		if err != nil {
			return nil, err
		}
		if len(pkt.Payload) >= 0x10000 {
			return nil, ErrMessageTooLarge
		}
		pkt.Length = uint16(len(pkt.Payload))
	} else {	
		pkt.Length = 0
	}

	return pkt, nil
}

type MsgNop struct {
}

func NewNop() (msg *MsgNop) {
	msg = new(MsgNop)

	return msg
}

func (msg *MsgNop) Type() (ptype uint8) {
	return TypeNop
}

func (msg *MsgNop) Marshal() (resp interface{}) {
	return nil
}

type MsgIdentifyClient struct {
	Hostname	string
	ProtocolVersion	uint32
}

func NewIdentifyClient() (msg *MsgIdentifyClient) {
	msg = new(MsgIdentifyClient)
	msg.ProtocolVersion = 1
	
	return msg
}

func (msg *MsgIdentifyClient) Type() (ptype uint8) {
	return TypeIdentifyClient
}

func (msg *MsgIdentifyClient) Marshal() (resp interface{}) {
	ps := new(ProtoIdentifyClient)

	ps.Hostname = proto.String(msg.Hostname)
	ps.ProtocolVersion = proto.Uint32(msg.ProtocolVersion)

	return ps
}