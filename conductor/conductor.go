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

var x509CertFilename *string = flag.String("cert", "conductor_crt.pem", "File containing the Certificate")
var x509PrivateKeyFilename *string = flag.String("key", "conductor_key.pem", "File containing the Private Key")
var bindAddress    *string = flag.String("bind-addr", "", "Bind Address")


func main() {
	var sockConfig tls.Config

	flag.Parse()

	var bindIp *net.IPAddr = nil
	if (*bindAddress != "") {
		var err os.Error
		bindIp, err = net.ResolveIPAddr(*bindAddress)
		if (err != nil) {
			o.Warn("Ignoring bind address.  Couldn't resolve \"%s\": %s", bindAddress, err)
		} else {
			bindIp = nil
		}
	}
	certpair, err := tls.LoadX509KeyPair(*x509CertFilename, *x509PrivateKeyFilename)
	o.MightFail("Couldn't load certificates", err)
	
	sockConfig.ServerName = o.ProbeHostname()

	StartRegistry()
	InitDispatch()
	StartHTTP()

	ServiceRequests(bindIp, nil, certpair)
}
