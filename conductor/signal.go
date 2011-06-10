/* signal.go
 *
 * Signal Handlers
*/

package main

import (
	"os/signal"
	"os"
	"syscall"
	"fmt"
	o "orchestra"
)


// handle the signals.  By default, we ignore everything, but the
// three terminal signals, HUP, INT, TERM, we want to explicitly
// handle.
func signalHandler() {
	for {
		sig := <- signal.Incoming

		ux, ok := sig.(signal.UnixSignal)
		if !ok {
			o.Warn("Couldn't handle signal %s, Coercion failed", sig)
			continue;
		}

		switch int(ux) {
		case syscall.SIGHUP:
			o.Warn("Reloading Configuration")
			//FIXME: Reload Config
		case syscall.SIGINT:
			fmt.Fprintln(os.Stderr, "Interrupt Received - Terminating")
			//FIXME: Gentle Shutdown
			os.Exit(1)
		case syscall.SIGTERM:
			fmt.Fprintln(os.Stderr, "Terminate Received - Terminating")
			//FIXME: Gentle Shutdown
			os.Exit(2)
		}
	}
	
}

func init() {
	go signalHandler()
}