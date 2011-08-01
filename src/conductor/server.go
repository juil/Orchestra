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
		o.MightFail(err, "Failed to get full hostname for bind address")
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
	o.Info("Binding to %s", laddr)
	listener, err := tls.Listen("tcp", laddr, &sockConfig)
	o.MightFail(err, "Couldn't bind TLS listener")

	for {
		o.Warn("Waiting for Connection...")
		c, err := listener.Accept()
		o.MightFail(err, "Couldn't accept TLS connection")
		o.Warn("Connection received from %s", c.RemoteAddr().String())
		HandleConnection(c)
	}
}



