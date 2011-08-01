/* various important shared defaults. */
package orchestra

import (
	"os"
	"net"
	"syslog"
	"fmt"
	"runtime/debug"
)

const (
	DefaultMasterPort = 2258
	DefaultHTTPPort = 2259
)

var	logWriter, _ = syslog.New(syslog.LOG_DEBUG, "orchestra")

func SetLogName(name string) {
	if nil != logWriter {
		logWriter.Close()
		logWriter = nil
	}
	var err os.Error
	logWriter, err = syslog.New(syslog.LOG_DEBUG, name)
	MightFail(err, "Couldn't reopen syslog")
}


func Debug(format string, args ...interface{}) {
	logWriter.Debug(fmt.Sprintf(format, args...))
}

func Info(format string, args ...interface{}) {
	logWriter.Info(fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	logWriter.Warning(fmt.Sprintf(format, args...))
}

func Fail(mesg string, args ...interface {}) {
	logWriter.Err(fmt.Sprintf(mesg, args...))
	fmt.Fprintf(os.Stderr, "ERR: "+mesg, args...);
	os.Exit(1)
}	

func MightFail(err os.Error, mesg string, args ...interface {}) {
	if (nil != err) {
		imesg := fmt.Sprintf(mesg, args...)
		Fail("%s: %s", imesg, err.String())
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
	MightFail(err, "Failed to get address for hostname")
	hostnames, err := net.LookupAddr(addr[0])
	MightFail(err, "Failed to get full hostname for address")

	return hostnames[0]
}
