/* player.go
*/

package main

import (
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
			go Reader(conn)

			/* Introduce ourself */
			p := o.NewIdentifyClient(*localHostname)			
			p.Send(conn)

			loop := true
			for loop {
				o.Warn("Waiting for some action!")
				select {
				case <-lostConnection:
					o.Warn("Lost Connection to Master")
					loop = false
				case p := <-receivedMessage:
					o.Warn("The Master spoke to me!")
					p.Dump()
				case <-time.After(KeepaliveDelay):
					o.Warn("Sending Nop")
					p := o.MakeNop()
					p.Send(conn)
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

	ProcessingLoop()
}
