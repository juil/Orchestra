// submitjob.go
//
// A sample Orchestra submission client.

package main

import (
	"io"
	"net"
	"json"
	"flag"
	"fmt"
	"os"
)

type JobRequest struct {
	Op	string
	Score	string
	Players	[]string
	Scope	string
	Params	map[string]string
}

var (
	AllOf	     = flag.Bool("all-of", false, "Send request to all named players")
	AudienceSock = flag.String("audience-sock", "/var/run/conductor.sock", "Path for the audience submission socket")
)

func NewJobRequest() (jr *JobRequest) {
	jr = new(JobRequest)
	jr.Params = make(map[string]string)

	return jr
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [<options>] <score> <player1> [<player2>...] [! [<key1> <value1>]...]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	jr := NewJobRequest()
	jr.Op = "queue"
	jr.Score = args[0]
	if *AllOf {
		jr.Scope = "all"
	} else {
		jr.Scope = "one"
	}

	var k int
	for k = 1; k < len(args); k++ {
		if args[k] == "!" {
			break;
		}
		insertionpoint := 0
		if nil == jr.Players {
			jr.Players = make([]string, 1)
		} else {
			insertionpoint = len(jr.Players)
			newplayers := make([]string, insertionpoint+1)
			copy(newplayers, jr.Players)
			jr.Players = newplayers
		}
		jr.Players[insertionpoint] = args[k]
	}
	if (k < len(args)) {
		// skip the !
		k++
		if (len(args) - (k))%2 != 0 {
			fmt.Fprintf(os.Stderr, "Error: Odd number of param arguments.\n")
			os.Exit(1)
		}
		for ; k < len(args); k+=2 {
			jr.Params[args[k]] = args[k+1]
		}
	}
	
	raddr, err := net.ResolveUnixAddr("unix", *AudienceSock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve sockaddr: %s\n", err)
		os.Exit(1)
	}
	conn, err := net.DialUnix("unix", nil, raddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to sockaddr: %s\n", err)
		os.Exit(1)
	}

	defer conn.Close()

	conn.SetTimeout(0)

	nc := net.Conn(conn)

	r, _ := nc.(io.Reader)
	w, _ := nc.(io.Writer)

	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)

	// send the message
	err = enc.Encode(jr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal & send: %s\n", err)
		os.Exit(1)
	}

	response := new([2]interface{})
	err = dec.Decode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %s\n", err)
		os.Exit(1)
	}
	// coerce field 0 back into a string.
	rerr,ok := response[0].(string)
	if ok {
		if rerr == "OK" {
			// all OK!  get the JobID
			jobid, _ := response[1].(float64)
			fmt.Printf("%d\n", uint64(jobid))
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "Server Error: %s\n", rerr)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Couldn't unmarshal response correctly.\n");
		os.Exit(1)
	}
}
