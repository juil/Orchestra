/* conductor.go
*/

package main

import (
	"os"
	"fmt"
	"net"
	"flag"
	"crypto/tls"
)

var x509CertFilename *string = flag.String("cert", "conductor_crt.pem", "File containing the Certificate")
var x509PrivateKeyFilename *string = flag.String("key", "conductor_key.pem", "File containing the Private Key")
var bindAddress    *string = flag.String("bind-addr", "", "Bind Address")

func Warn(format string, args ...interface{}) {
	fmt.Fprint(os.Stderr, "WARN: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
}

func Fail(mesg string, args ...interface {}) {
	fmt.Fprint(os.Stderr, "ERR: ")
	fmt.Fprintf(os.Stderr, mesg, args...);
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}	

func MightFail(mesg string, err os.Error) {
	if (nil != err) {
		Fail("%s: %s", mesg, err.String())
	}
}

func main() {
	var sockConfig tls.Config

	flag.Parse()

	var bindIp *net.IPAddr = nil
	if (*bindAddress != "") {
		var err os.Error
		bindIp, err = net.ResolveIPAddr(*bindAddress)
		if (err != nil) {
			Warn("Ignoring bind address.  Couldn't resolve \"%s\": %s", bindAddress, err)
		} else {
			bindIp = nil
		}
	}
	certpair, err := tls.LoadX509KeyPair(*x509CertFilename, *x509PrivateKeyFilename)
	MightFail("Couldn't load certificates", err)
	SetupMasterSocket(bindIp, nil, certpair)
	
	sockConfig.ServerName = ProbeHostname()

	InitDispatch()

	waitingT, _ := DispatchStatus()

	fmt.Printf("Hostname: %s\n", sockConfig.ServerName)
	fmt.Printf("Tasks Waiting: %d\n", waitingT)
	fmt.Printf("Got Here OK\n")
}
