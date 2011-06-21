// getstatus.go
//
// A sample Orchestra status polling client.

package main

import (
	"io"
	"net"
	"json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type StatusRequest struct {
	Op	string
	Id	uint64
}

type PlayerStatus struct {
	Status		*string
	Response	map[string]*string
}

type StatusResponse struct {
	Status		*string
	Players		map[string]*PlayerStatus
}

var (
	AudienceSock = flag.String("audience-sock", "/var/run/orchestra/audience.sock", "Path for the audience submission socket")
)

func NewStatusRequest() (sr *StatusRequest) {
	sr = new(StatusRequest)
	sr.Op = "status"
	return sr
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [<options>] <jobid>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	sr := NewStatusRequest()
	var err os.Error
	sr.Id, err = strconv.Atoui64(flag.Arg(0))
	if nil != err {
		fmt.Fprintf(os.Stderr, "Failed to parse JobID: %s\n", err)
		os.Exit(1)
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
	err = enc.Encode(sr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal & send: %s\n", err)
		os.Exit(1)
	}

	response := new([2]interface{})
	sresp := new(StatusResponse)
	response[1] = sresp
	err = dec.Decode(response)
	if err != nil {
		// OK, the problem here is that an in the request will throw a null in the second field.
		// This will cause Decode to softfault.  We'll ignore these softfaults.
		utye, ok := err.(*json.UnmarshalTypeError)
		if ok {
			fmt.Fprintf(os.Stderr, "Unmarshalling error: %s of Type %s\n", utye.Value, utye.Type)
		} else {
			ufe, ok := err.(*json.UnmarshalFieldError)
			if ok {
				fmt.Fprintf(os.Stderr, "Error decoding response: UFE %s of Type %s\n", ufe.Key, ufe.Type)
				os.Exit(1)
			}
			ute, ok := err.(*json.UnsupportedTypeError)
			if ok {
				fmt.Fprintf(os.Stderr, "Error decoding response: UTE of Type %s\n", ute.Type)
				os.Exit(1)
			}

			fmt.Fprintf(os.Stderr, "Error decoding response: %s\n", err)
			os.Exit(1)
		}
	}

	// coerce field 0 back into a string.
	rerr,ok := response[0].(string)
	if ok {
		if rerr == "OK" {
			// all OK, process the sresp.
			fmt.Printf("Aggregate: %s\n", *sresp.Status)
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
