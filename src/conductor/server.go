/* server.go
 *
 * TLS and Connection Hell.
*/

package main

import (
	"net"
	"crypto/tls"
	"fmt"
	"os"
	"crypto/x509"
	o	"orchestra"
)


func ServiceRequests() {
	var sockConfig tls.Config

	// resolve the bind address
	bindAddressStr := GetStringOpt("bind address")
	var bindAddr *net.IPAddr = nil
	if (bindAddressStr != "") {
		var err os.Error
		bindAddr, err = net.ResolveIPAddr("ip", bindAddressStr)
		if (err != nil) {
			o.Warn("Ignoring bind address.  Couldn't resolve \"%s\": %s", bindAddressStr, err)
		} else {
			bindAddr = nil
		}
	}
	// load the x509 certificate and key, then attach it to the tls config.
	x509CertFilename := GetStringOpt("x509 certificate")
	x509PrivateKeyFilename := GetStringOpt("x509 private key")
	serverCert, err := tls.LoadX509KeyPair(x509CertFilename, x509PrivateKeyFilename)
	o.MightFail(err, "Couldn't load certificates")
	sockConfig.Certificates = append(sockConfig.Certificates, serverCert)

	// load the CA certs
	caCertPool := x509.NewCertPool()
	caCertNames := GetCACertList()
	if caCertNames != nil {
		for _, filename := range caCertNames {
			fh, err := os.Open(filename)
			if err != nil {
				o.Warn("Whilst parsing CA certs, couldn't open %s: %s", filename, err)
				continue
			}
			defer fh.Close()
			fi, err := fh.Stat()
			o.MightFail(err, "Couldn't stat CA certificate file: %s", filename)
			data := make([]byte, fi.Size)
			fh.Read(data)
			caCertPool.AppendCertsFromPEM(data)
		}
	}
	sockConfig.RootCAs = caCertPool

	// determine the server hostname.
	servername := GetStringOpt("server name")
	if servername != "" {
		o.Info("Using %s as the server name", servername)
		sockConfig.ServerName = servername
	} else {
		if bindAddr != nil {
			o.Warn("Probing for fqdn for bind address as none was provided.")
			hostnames, err := net.LookupAddr(bindAddr.String())
			o.MightFail(err, "Failed to get full hostname for bind address")
			sockConfig.ServerName = hostnames[0]
		} else {
			o.Warn("Probing for fqdn as no server name was provided")
			sockConfig.ServerName = o.ProbeHostname()
		}
	}

	// ask the client to authenticate
	sockConfig.AuthenticateClient = true

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



