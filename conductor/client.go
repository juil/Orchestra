/* client.go
 *
 * Client Handling
*/

package main
import (
	o "orchestra"
	"net"
	"time"
	"os"
)

const (
//	KeepaliveDelay =	300e9 // once every 5 minutes.
	KeepaliveDelay =	5e9 // once every 5 seconds for debug
	OutputQueueDepth =	10
)


type ClientInfo struct {
	Hostname	string
	PktOutQ		chan *o.WirePkt
	PktInQ		chan *o.WirePkt
	AbortQ		chan int
	TaskQ		chan *JobTask
	connection	net.Conn	
}

func NewClientInfo() (client *ClientInfo) {
	client = new(ClientInfo)
	client.AbortQ = make(chan int, 2)
	client.PktOutQ = make(chan *o.WirePkt, OutputQueueDepth)
	client.PktInQ = make(chan *o.WirePkt)

	return client
}

func (client *ClientInfo) Abort() {
	client.AbortQ <- 1;
}

func (client *ClientInfo) Name() (name string) {
	if client.Hostname == "" {
		return "UNK:" + client.connection.RemoteAddr().String()
	}
	return client.Hostname
}


func handleNop(client *ClientInfo, message interface{}) {
	o.Warn("Client %s: NOP Received", client.Name())
}

func handleIdentify(client *ClientInfo, message interface{}) {
	if client.Hostname != "" {
		o.Warn("Client %s: Tried to reintroduce itself.", client.Name())
		client.Abort()
		return
	}
	ic, _ := message.(*o.IdentifyClient)
	o.Warn("Client %s: Identified Itself As \"%s\"", client.Name(), *ic.Hostname)
	client.Hostname = *ic.Hostname
	if !ClientAdd(client.Hostname, client) {
		o.Warn("Couldn't register client %s.  aborting connection.", client.Name())
		client.Abort()
	}
}

func handleReadyForTask(client *ClientInfo, message interface{}) {
	o.Warn("Client %s: Asked for Job", client.Name())
	PlayerWaitingForJob(client)
}

func handleIllegal(client *ClientInfo, message interface{}) {
	o.Warn("Client %s: Sent Illegal Message")
	client.Abort()
}

var dispatcher	= map[uint8] func(*ClientInfo,interface{}) {
	o.TypeNop:		handleNop,
	o.TypeIdentifyClient:	handleIdentify,
	o.TypeReadyForTask:	handleReadyForTask,

	/* C->P only messages, should never appear on the wire. */
	o.TypeTaskRequest:	handleIllegal,

}

func clientLogic(client *ClientInfo) {
	loop := true
	for loop {
		o.Warn("CL:%s:Select", client.Name())
		select {
		case p := <-client.PktInQ:
			/* we've received a packet.  do something with it. */
			if client.Hostname == "" && p.Type != o.TypeIdentifyClient {
				o.Warn("Client %s didn't Identify self - got type %d instead!", client.Name(), p.Type)
				client.Abort()
				break
			}
			var upkt interface {} = nil
			if p.Length > 0 {
				var err os.Error

				upkt, err = p.Decode()
				if err != nil {
					o.Warn("Error unmarshalling message from Client %s: %s", client.Name(), err)
					client.Abort()
					break
				}
			}
			handler, exists := dispatcher[p.Type]
			if (exists) {
				handler(client, upkt)
			} else {
				o.Warn("Unhandled Pkt Type %d", p.Type)
			}
		case p := <-client.PktOutQ:
			if p != nil {
				_, err := p.Send(client.connection)
				if err != nil {
					o.Warn("Error sending pkt to %s: %s", client.Name(), err)
					client.Abort()
				}
			}
		case <-client.AbortQ:
			o.Warn("Client %s connection writer on fire!", client.Name())
			PlayerDied(client)
			ClientDelete(client.Hostname)
			loop = false

		case <-time.After(KeepaliveDelay):
			p := o.MakeNop()
			o.Warn("Sending Keepalive to %s", client.Name())
			_, err := p.Send(client.connection)
			if err != nil {
				o.Warn("Error sending pkt to %s: %s", client.Name(), err)	
				client.Abort()
			}
		}
	}
	client.connection.Close()
}

func clientReceiver(client *ClientInfo) {
	loop := true
	for loop {
		pkt, err := o.Receive(client.connection)
		if nil != err {
			o.Warn("Error receiving pkt from host %s: %s", client.Name(), err)
			client.Abort()
			client.connection.Close()
			loop = false
		} else {
			client.PktInQ <- pkt
		}
	}
	o.Warn("Client %s connection reader on fire!", client.Name())
}

/* The Main Server loop calls this method to hand off connections to us */
func HandleConnection(conn net.Conn) {
	c := NewClientInfo()
	c.connection = conn
	
	go clientReceiver(c)
	go clientLogic(c)
}