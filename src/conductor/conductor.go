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
	x509CertFilename = flag.String("cert", "conductor_crt.pem", "File containing the Certificate")
	x509PrivateKeyFilename = flag.String("key", "conductor_key.pem", "File containing the Private Key")
	bindAddress = flag.String("bind-addr", "", "Bind Address")
	ConfigDirectory = flag.String("config-dir", "/etc/conductor", "Configuration Directory")
	AudienceSock = flag.String("audience-sock", "/var/run/conductor.sock", "Path for the audience submission socket")
	StateDir = flag.String("state-dir", "/var/spool/orchestra", "State/Spool directory for storing persistant state between runs")
)


func main() {
	var sockConfig tls.Config

	flag.Parse()

	var bindIp *net.IPAddr = nil
	if (*bindAddress != "") {
		var err os.Error
		bindIp, err = net.ResolveIPAddr("ip", *bindAddress)
		if (err != nil) {
			o.Warn("Ignoring bind address.  Couldn't resolve \"%s\": %s", bindAddress, err)
		} else {
			bindIp = nil
		}
	}
	certpair, err := tls.LoadX509KeyPair(*x509CertFilename, *x509PrivateKeyFilename)
	o.MightFail("Couldn't load certificates", err)
	
	sockConfig.ServerName = o.ProbeHostname()

	// Start the client registry - configuration parsing will block indefinately
	// if the registry listener isn't working
	StartRegistry()
	// do an initial configuration load
	ConfigLoad()

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
