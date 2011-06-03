/* server.go
 *
 * TLS and Connection Hell.
*/

package main

import (
	"os"
	"net"
	"crypto/tls"
	"fmt"
)

const (
	DefaultMasterPort = 2258
	DefaultHTTPPort = 2259
)

func ProbeHostname() (fqdn string) {
	var shortHostname string

	shortHostname, err := os.Hostname()
	addr, err := net.LookupHost(shortHostname) 
	MightFail("Failed to get address for hostname", err)
	hostnames, err := net.LookupAddr(addr[0])
	MightFail("Failed to get full hostname for address", err)

	return hostnames[0]
}

func ServiceRequests(bindAddr *net.IPAddr, hostname *string, serverCert tls.Certificate) {
	var sockConfig tls.Config

	/* we have a bindAddr and validate it */
	if (bindAddr != nil && hostname == nil) {
		Warn("Probing for fqdn for bind address as none was provided.")
		hostnames, err := net.LookupAddr(bindAddr.String())
		MightFail("Failed to get full hostname for bind address", err)
		sockConfig.ServerName = hostnames[0]
	} else {
		if (hostname != nil) {
			sockConfig.ServerName = *hostname
		} else {
			if (hostname == nil) {
				Warn("Probing for fqdn as none was provided.")
				sockConfig.ServerName = ProbeHostname()
			}
		}
	}
	/* attach the certificate chain! */
	sockConfig.Certificates = make([]tls.Certificate, 1)
	sockConfig.Certificates[0] = serverCert

	/* convert the bindAddress to a string suitable for the Listen call */
	var laddr string
	if (bindAddr == nil) {
		laddr = fmt.Sprintf(":%d", DefaultMasterPort)
	} else {
		laddr = fmt.Sprintf("%s:%d", bindAddr.String(), DefaultMasterPort)
	}
	Warn("Binding to %s", laddr)
	listener, err := tls.Listen("tcp", laddr, &sockConfig)
	MightFail("Couldn't bind TLS listener", err)

	for {
		c, err := listener.Accept()
		MightFail("Couldn't accept TLS connection", err)
		Warn("Connection received from %s", c.RemoteAddr().String())
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	// escalte to tls Conn if we can.
	/*FIXME: implement client registraiton, sender + receive loop */
	c.Close()
}