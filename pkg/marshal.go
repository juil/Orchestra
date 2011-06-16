/* marshal.go
 *
 * Common wire marshalling code.
*/

package orchestra;

import (
	"os"
	"goprotobuf.googlecode.com/hg/proto"
)

var (
	ErrUnknownType = os.NewError("Unknown Type in Encode request")
	ErrObjectTooLarge = os.NewError("Encoded Object exceeds maximum encoding size")
)

func (p *WirePkt) Decode() (obj interface{}, err os.Error) {
	switch (p.Type) {
	case TypeNop:
		if (p.Length != 0) {
			/* throw error later... */
			return nil, ErrMalformedMessage;
		}
		return nil, nil
	case TypeIdentifyClient:
		ic := new(IdentifyClient)
		err := proto.Unmarshal(p.Payload[0:p.Length], ic)
		if err != nil {
			return nil, err
		}
		return ic, nil
	case TypeReadyForTask:
		if (p.Length != 0) {
			/* throw error later... */
			return nil, ErrMalformedMessage;
		}
		return nil, nil
	case TypeTaskRequest:
		tr := new(ProtoTaskRequest)
		err := proto.Unmarshal(p.Payload[0:p.Length], tr)
		if err != nil {
			return nil, err
		}
		return tr, nil
	case TypeTaskResponse:
		tr := new(ProtoTaskResponse)
		err := proto.Unmarshal(p.Payload[0:p.Length], tr)
		if err != nil {
			return nil, err
		}
		return tr, nil
	case TypeAcknowledgement:
		tr := new(ProtoAcknowledgement)
		err := proto.Unmarshal(p.Payload[0:p.Length], tr)
		if err != nil {
			return nil, err
		}
		return tr, nil
	}
	return nil, ErrUnknownMessage
}

func Encode(obj interface{}) (p *WirePkt, err os.Error) {
	p = new(WirePkt)
	switch obj.(type) {
	case *IdentifyClient:
		p.Type = TypeIdentifyClient
	case *ProtoTaskRequest:
		p.Type = TypeTaskRequest
	case *ProtoTaskResponse:
		p.Type = TypeTaskResponse
	case *ProtoAcknowledgement:
		p.Type = TypeAcknowledgement
	default:
		Warn("Encoding unknown type!")
		return nil, ErrUnknownType
	}
	p.Payload, err = proto.Marshal(obj)
	if err != nil {
		return nil, err
	}
	if len(p.Payload) >= 0x10000 {
		return nil, ErrObjectTooLarge
	}
	p.Length = uint16(len(p.Payload))

	return p, nil	
}


func MakeNop() (p *WirePkt) {
	p = new(WirePkt)
	p.Length = 0
	p.Type = TypeNop
	p.Payload = nil

	return p
}

func MakeIdentifyClient(hostname string) (p *WirePkt) {
	s := new(IdentifyClient)
	s.Hostname = proto.String(hostname)

	p, _ = Encode(s)
	
	return p
}

func MakeReadyForTask() (p *WirePkt){
	p = new(WirePkt)
	p.Type = TypeReadyForTask
	p.Length = 0
	p.Payload = nil

	return p
}