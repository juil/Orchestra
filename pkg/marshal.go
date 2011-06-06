/* marshal.go
 *
 * Common wire marshalling code.
*/

package orchestra;

import (
	"os"
	"goprotobuf.googlecode.com/hg/proto"
)

func (p *WirePkt) Decode() (obj interface{}, err os.Error) {
	switch (p.Type) {
	case TypeNop:
		if (p.Length != 0) {
			/* throw error later... */
			return nil, ErrMalformedMessage;
		}
	case TypeIdentifyClient:
		ic := new(ProtoIdentifyClient)
		err := proto.Unmarshal(p.Payload[0:p.Length], ic)
		if err != nil {
			return nil, err
		}
		return ic, nil
	case TypeTaskRequest:
		tr := new(ProtoTaskRequest)
		err := proto.Unmarshal(p.Payload[0:p.Length], tr)
		if err != nil {
			return nil, err
		}
		return tr, nil
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
	s := new(ProtoIdentifyClient)
	s.Hostname = proto.String(hostname)
	
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
	p.Type = TypeIdentifyClient

	return p
}
