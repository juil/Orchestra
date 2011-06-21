/* audience.go
*/

package main

import (
	"io"
	"json"
	"net"
	o "orchestra"
)

type GenericJsonRequest struct {
	Op		*string
	Score		*string
	Players		[]string
	Scope		*string
	Params		map[string]string
	Id		*uint64
}

func handleAudienceRequest(c net.Conn) {
	defer c.Close()

	c.SetTimeout(0)
	r, _ := c.(io.Reader)
	w, _ := c.(io.Writer)
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)

	outobj := new(GenericJsonRequest)
	err := dec.Decode(outobj)
	if err != nil {
		o.Warn("Error decoding JSON talking to audience: %s", err)
		return
	}

	if nil == outobj.Op {
		o.Warn("Malformed JSON message talking to audience.  Missing Op")
		return
	}
	switch *(outobj.Op) {
	case "status":
		o.Warn("Status...")
	case "queue":
		if nil == outobj.Score {
			o.Warn("Malformed Queue message talking to audience. Missing Score")
			return
		}
		if nil == outobj.Scope {
			o.Warn("Malformed Queue message talking to audience. Missing Scope")
			return
		}
		if nil == outobj.Players || len(outobj.Players) < 1 {
			o.Warn("Malformed Queue message talking to audience. Missing Players")
			return
		}
		job := NewRequest()
		job.Score = *outobj.Score
		switch (*outobj.Scope) {
		case "one":
			job.Scope = o.SCOPE_ONEOF
		case "all":
			job.Scope = o.SCOPE_ALLOF
		}
		job.Players = outobj.Players
		job.Params = outobj.Params
		QueueJob(job)
		/* build response */
		resp := make([]interface{},2)
		resperr := new(string)
		*resperr = "OK"
		resp[0] = resperr

		// this probably looks odd, but all numbers cross through float64 when being json encoded.  d'oh!
		jobid := new(uint64)
		*jobid = uint64(job.Id)
		resp[1] = jobid

		err := enc.Encode(resp)
		if nil != err {
			o.Warn("Couldn't encode response to audience: %s", err)
		}
	default:
		o.Warn("Unknown operation talking to audience: \"%s\"", *(outobj.Op))
		return
	}

	_ = enc
}

func NewAudienceListener(l net.Listener) {
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			o.Warn("Accept() failed on Audience Listenter.")
			break
		}
		go handleAudienceRequest(c)
	}
}

func StartAudienceSock() {
	laddr, err := net.ResolveUnixAddr("unix", *AudienceSock)
	o.MightFail("Couldn't resolve audience socket address", err)
	l, err := net.ListenUnix("unix", laddr)
	o.MightFail("Couldn't start audience unixsock listener", err)
	go NewAudienceListener(l)	
}