/* various important shared defaults. */
package orchestra

import (
	"os"
	"net"
	"log"
	"fmt"
	"runtime/debug"
)

const (
	DefaultMasterPort = 2258
	DefaultHTTPPort = 2259
)

func Warn(format string, args ...interface{}) {
	log.Printf("WARN: "+format, args...)
}

func Fail(mesg string, args ...interface {}) {
	log.Fatalf("ERR: "+mesg, args...);
}	

func MightFail(mesg string, err os.Error) {
	if (nil != err) {
		Fail("%s: %s", mesg, err.String())
	}
}


// Throws a generic assertion error, stacktraces, dies.
// only really to be used where the runtime-time configuration
// fucks up internally, not for user induced errors.
func Assert(mesg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, mesg, args...)
	debug.PrintStack()
	os.Exit(1)
}

func ProbeHostname() (fqdn string) {
	var shortHostname string

	shortHostname, err := os.Hostname()
	addr, err := net.LookupHost(shortHostname)
	MightFail("Failed to get address for hostname", err)
	hostnames, err := net.LookupAddr(addr[0])
	MightFail("Failed to get full hostname for address", err)

	return hostnames[0]
}
