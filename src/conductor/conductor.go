/* conductor.go
*/

package main

import (
	"os"
	"net"
	"flag"
	"crypto/tls"
	o	"orchestra"
)

var (
	ConfigFile = flag.String("config-file", "/etc/conductor/conductor.conf", "File containing the conductor configuration")
)


func main() {
	var sockConfig tls.Config

	// parse command line options.
	flag.Parse()

	// Start the client registry - configuration parsing will block indefinately
	// if the registry listener isn't working
	StartRegistry()
	// do an initial configuration load
	ConfigLoad()

	// set up stuff now.
	bindAddress := GetStringOpt("bind address")
	var bindIp *net.IPAddr = nil
	if (bindAddress != "") {
		var err os.Error
		bindIp, err = net.ResolveIPAddr("ip", bindAddress)
		if (err != nil) {
			o.Warn("Ignoring bind address.  Couldn't resolve \"%s\": %s", bindAddress, err)
		} else {
			bindIp = nil
		}
	}

	x509CertFilename := GetStringOpt("x509 certificate")
	x509PrivateKeyFilename := GetStringOpt("x509 private key")
	certpair, err := tls.LoadX509KeyPair(x509CertFilename, x509PrivateKeyFilename)
	o.MightFail(err, "Couldn't load certificates")

	sockConfig.ServerName = GetStringOpt("server name")
	if sockConfig.ServerName == "" {
		sockConfig.ServerName = o.ProbeHostname()
	}

	// start the master dispatch system
	InitDispatch()
	defer CleanDispatch()

	// start the status listener
	StartHTTP()
	// start the audience listener
	StartAudienceSock()
	// service TLS requests.
	ServiceRequests(bindIp, nil, certpair)
}
