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
	"goprotobuf.googlecode.com/hg/proto"
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
	// escalte to tls Conn if we can.
	pkt, err := o.Receive(c)
	if err != nil {
		o.Warn("Client didn't complete introduction: %s", err)
	} else {
		o.Warn("Got Pkt.  Type=%d.  Len=%d", pkt.Type, pkt.Length)
		if pkt.Type != o.TypeIdentifyClient {
			o.Warn("Client didn't Identify self! (Got PktType %d)", pkt.Type)
		} else {
			upkt, err := pkt.Decode()
			if err != nil {
				o.Warn("Error unmarshalling Introduction: %s", err)
			} else {
				ic, _ := upkt.(o.ProtoIdentifyClient)
				if (ic.Hostname == nil) {
					o.Fail("Couldn't find hostname in Identity packet!")
				} else {
					remoteHostname = proto.GetString(ic.Hostname)
				}
			}
		}
	}
	o.Warn("Client Identified Itself As \"%s\"", remoteHostname)
	/*FIXME: implement client registraiton, sender + receive loop */
	c.Close()
}