/* server.go
 *
 * TLS and Connection Hell.
*/

package main

import (
	"net"
	"crypto/tls"
	"fmt"
	o	"orchestra"
)


func ServiceRequests(bindAddr *net.IPAddr, hostname *string, serverCert tls.Certificate) {
	var sockConfig tls.Config

	/* we have a bindAddr and validate it */
	if (bindAddr != nil && hostname == nil) {
		o.Warn("Probing for fqdn for bind address as none was provided.")
		hostnames, err := net.LookupAddr(bindAddr.String())
		o.MightFail("Failed to get full hostname for bind address", err)
		sockConfig.ServerName = hostnames[0]
	} else {
		if (hostname != nil) {
			sockConfig.ServerName = *hostname
		} else {
			if (hostname == nil) {
				o.Warn("Probing for fqdn as none was provided.")
				sockConfig.ServerName = o.ProbeHostname()
			}
		}
	}
	/* attach the certificate chain! */
	sockConfig.Certificates = make([]tls.Certificate, 1)
	sockConfig.Certificates[0] = serverCert

	/* convert the bindAddress to a string suitable for the Listen call */
	var laddr string
	if (bindAddr == nil) {
		laddr = fmt.Sprintf(":%d", o.DefaultMasterPort)
	} else {
		laddr = fmt.Sprintf("%s:%d", bindAddr.String(), o.DefaultMasterPort)
	}
	o.Warn("Binding to %s", laddr)
	listener, err := tls.Listen("tcp", laddr, &sockConfig)
	o.MightFail("Couldn't bind TLS listener", err)

	for {
		c, err := listener.Accept()
		o.MightFail("Couldn't accept TLS connection", err)
		o.Warn("Connection received from %s", c.RemoteAddr().String())
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	var remoteHostname string = ""

	for {
		pkt, err := o.Receive(c)
		if err != nil {
			o.Warn("Error receiving message: %s", err)
			break
		}
		if remoteHostname == "" && pkt.Type != o.TypeIdentifyClient {
			o.Warn("Client didn't Identify self - got type %d instead!", pkt.Type)
			break
		}
		upkt, err := pkt.Decode()
		if err != nil {
			o.Warn("Error unmarshalling message: %s", err)
		}
		switch pkt.Type {
		case o.TypeIdentifyClient:
			ic, _ := upkt.(*o.IdentifyClient)
			remoteHostname = *ic.Hostname
			o.Warn("Client Identified Itself As \"%s\"", remoteHostname)
		default:
			o.Warn("Unhandled Pkt Type %d", pkt.Type)
		}
	}
	/*FIXME: Sever the client from the clients list. */
	c.Close()
}