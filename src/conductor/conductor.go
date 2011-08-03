/* conductor.go
*/

package main

import (
	"flag"
	o	"orchestra"
)

var (
	ConfigFile = flag.String("config-file", "/etc/conductor/conductor.conf", "File containing the conductor configuration")
)


func main() {
	o.SetLogName("conductor")

	// parse command line options.
	flag.Parse()

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
	ServiceRequests()
}
