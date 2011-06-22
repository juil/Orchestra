/* player.go
*/

package main

import (
	"os"
	"fmt"
	"flag"
	o	"orchestra"
	"crypto/tls"
	"net"
	"time"
)

const (
	KeepaliveDelay = 5e9
)

var (
	masterHostname = flag.String("master", "conductor", "Hostname of the Conductor")
	localHostname  = flag.String("hostname", "", "My Hostname (defaults to autoprobing)")
	masterPort     = flag.Int("port", o.DefaultMasterPort, "Port to contact conductor on")
	ScoreDirectory = flag.String("score-dir", "/usr/lib/orchestra/scores", "Path for the Directory containing valid scores")
 	receivedMessage = make(chan *o.WirePkt)
	lostConnection = make(chan int)
)

func Reader(conn net.Conn) {
	for {
		pkt, err := o.Receive(conn)
		if (err != nil) {
			o.Warn("Error receiving message: %s", err)
			break;
		}
		receivedMessage <- pkt
	}
	lostConnection <- 1
	
}

func handleNop(c net.Conn, message interface{}) {
	o.Warn("NOP Received")
}

func handleIllegal(c net.Conn, message interface{}) {
	o.Fail("Got Illegal Message")
}

func handleRequest(c net.Conn, message interface{}) {
	o.Warn("Request Recieved.  Decoding!")
	ptr, ok := message.(*o.ProtoTaskRequest)
	if !ok {
		o.Fail("CC stuffed up - handleRequest got something that wasn't a ProtoTaskRequest.")
	}
	job := o.JobFromProto(ptr)
	/* search the registry for the job */
	o.Warn("Request for Job.ID %d", job.Id)
	existing := o.JobGet(job.Id)
	if nil != existing {
		if (existing.MyResponse.IsFinished()) {
			//FIXME: update retry time on Response
			ptr := existing.MyResponse.Encode()
			p, err := o.Encode(ptr)
			o.MightFail("Failed to encode response", err)
			_, err = p.Send(c)
			if err != nil {
				o.Warn("Transmission error: %s", err)
				c.Close()
				lostConnection <- 1
			}
		}
	} else {
		/* check to see if we have the score */		
		/* add the Job to our Registry */
		job.MyResponse = o.NewTaskResponse()
		job.MyResponse.State = o.RESP_PENDING		
		o.JobAdd(job)
		o.Warn("Added Job %d to our local registry", job.Id)
	}
}


var dispatcher	= map[uint8] func(net.Conn, interface{}) {
	o.TypeNop:		handleNop,
	o.TypeTaskRequest:	handleRequest,

	/* P->C only messages, should never appear on the wire to us. */
	o.TypeIdentifyClient:	handleIllegal,
	o.TypeReadyForTask:	handleIllegal,
	o.TypeTaskResponse:	handleIllegal,
}

func ProcessingLoop() {
	tconf := &tls.Config{
	}
	raddr := fmt.Sprintf("%s:%d", *masterHostname, *masterPort)

	for {
		o.Warn("Connecting to %s", raddr)
		conn, err := tls.Dial("tcp", raddr, tconf)
		if err != nil {
			o.Warn("Couldn't connect to master: %s", err)
			o.Warn("Sleeping for 30 seconds")
			err := time.Sleep(30 * 10e9)
			o.MightFail("Couldn't Sleep",err)
		} else {
			var pendingTaskRequest = false
			go Reader(conn)

			/* Introduce ourself */
			p := o.MakeIdentifyClient(*localHostname)
			p.Send(conn)

			loop := true
			for loop {
				o.Warn("Waiting for some action!")
				select {
				case <-lostConnection:
					o.Warn("Lost Connection to Master")
					loop = false
				case p := <-receivedMessage:
					var upkt interface{} = nil
					if p.Length > 0 {
						var err os.Error
						upkt, err = p.Decode()
						o.MightFail("Couldn't decode packet from master", err)
					}
					handler, exists := dispatcher[p.Type]
					if (exists) {
						handler(conn, upkt)
					} else {
						o.Fail("Unhandled Pkt Type %d", p.Type)
					}
				case <-time.After(KeepaliveDelay):
					if !pendingTaskRequest {
						o.Warn("Asking for trouble")
						p := o.MakeReadyForTask()
						p.Send(conn)
						o.Warn("Sent Request for trouble")
						pendingTaskRequest = true
					} else {
						o.Warn("Sending Nop")
						p := o.MakeNop()
						p.Send(conn)
					}
				}
			}
		}
	}
}

func main() {
	flag.Parse()

	if (*localHostname == "") {
		*localHostname = o.ProbeHostname()
		o.Warn("No hostname provided - probed hostname: %s", *localHostname)
	}
	LoadScores()
	ProcessingLoop()
}
