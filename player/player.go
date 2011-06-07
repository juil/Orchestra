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

var masterHostname *string = flag.String("master", "conductor", "Hostname of the Conductor")
var localHostname  *string = flag.String("hostname", "", "My Hostname (defaults to autoprobing)")
var masterPort     *int    = flag.Int("port", o.DefaultMasterPort, "Port to contact conductor on")

var receivedMessage = make(chan o.WirePkt)
var lostConnection = make(chan int)

func Reader(conn net.Conn) {
	for {
		pkt, err := o.Receive(conn)
		if (err != nil) {
			o.Warn("Error receiving message: %s", err)
			break;
		}
		receivedMessage <- *pkt
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

			breakLoop := true
			for breakLoop {
				o.Warn("Waiting for some action!")
				select {
				case <-lostConnection:
					o.Warn("Lost Connection to Master")
					breakLoop = false
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
