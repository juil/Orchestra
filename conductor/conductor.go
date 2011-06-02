/* conductor.go
*/

package main

import (
	"os"
	"fmt"
	"net"
	"flag"
	tls	"crypto/tls"
)


var rootCAFilename *string = flag.String("root-ca", "root-ca.pem", "File containing the CA Certificate")
var myCertFilename *string = flag.String("cert", "cert.pem", "File containing the Server Key and Certificate pair")
var bindAddress    *string = flag.String("bind-addr", "", "Bind Address")

func main() {
	var sockConfig tls.Config

	flag.Parse()

	var shortHostname string

	if ("" == *bindAddress) {
		var err os.Error
		shortHostname, err = os.Hostname()
		if (err != nil) {
			fmt.Printf("Failed to get hostname: %s\n", err);
			os.Exit(1);
		}
	} else {
		shortHostname = *bindAddress
	}
	addr, err := net.LookupHost(shortHostname) 
	if (err != nil) {
		fmt.Printf("Failed to get address for hostname: %s\n", err);
		os.Exit(1);
	}
	hostnames, err := net.LookupAddr(addr[0])
	if (err != nil) {
		fmt.Printf("Failed to get full hostname for address: %s\n", err);
		os.Exit(1);
	}

	sockConfig.ServerName = hostnames[0]

	InitDispatch()

	waitingT, _ := DispatchStatus()

	fmt.Printf("Hostname: %s\n", sockConfig.ServerName)
	fmt.Printf("Tasks Waiting: %d\n", waitingT)
	fmt.Printf("Got Here OK\n")
}
