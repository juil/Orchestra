/* audience.go
*/

package main

import (
	"io"
	"json"
	"net"
	"os"
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

type JsonPlayerStatus struct {
	Status		string
	Response	map[string]string
}

type JsonStatusResponse struct {
	Status		string
	Players		map[string]*JsonPlayerStatus
}

func NewJsonStatusResponse() (jsr *JsonStatusResponse) {
	jsr = new(JsonStatusResponse)
	jsr.Players = make(map[string]*JsonPlayerStatus)
	
	return jsr
}

func NewJsonPlayerStatus() (jps *JsonPlayerStatus) {
	jps = new(JsonPlayerStatus)
	jps.Response = make(map[string]string)

	return jps	
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
		if nil == outobj.Id {
			o.Warn("Malformed Status message talking to audience. Missing Job ID")
			return
		}
		job := o.JobGet(*outobj.Id)
		jresp := new([2]interface{})
		if nil != job {
			jresp[0] = "OK"
			iresp := NewJsonStatusResponse()
			switch job.State {
			case o.JOB_PENDING:
				iresp.Status = "PENDING"
			case o.JOB_SUCCESSFUL:
				iresp.Status = "OK"
			case o.JOB_FAILED_PARTIAL:
				iresp.Status = "PARTIAL_FAIL"
			case o.JOB_FAILED:
				iresp.Status = "FAIL"
			default:
				o.Fail("Blargh.  %d is an unknown job state!", job.State)
			}
			resnames := o.JobGetResultNames(*outobj.Id)
			for i := range resnames {
				tr := o.JobGetResult(*outobj.Id, resnames[i])
				if nil != tr {
					presp := NewJsonPlayerStatus()
					switch tr.State {
					case o.RESP_RUNNING:
						presp.Status = "PENDING"
					case o.RESP_FINISHED:
						presp.Status = "OK"
					case o.RESP_FAILED:
						presp.Status = "FAIL"
					case o.RESP_FAILED_UNKNOWN_SCORE:
						presp.Status = "UNK_SCORE"
					case o.RESP_FAILED_HOST_ERROR:
						presp.Status = "HOST_ERROR"
					case o.RESP_FAILED_UNKNOWN:
						presp.Status = "UNKNOWN_FAILURE"
					}
					for k,v:=range(tr.Response) {
						presp.Response[k] = v
					}
					iresp.Players[resnames[i]] = presp
				}
		
			}
			jresp[1] = iresp
		} else {
			jresp[0] = "Error"
			jresp[1] = nil
		}
		enc.Encode(jresp)
		o.Warn("Status...")
	case "queue":
		if nil == outobj.Score {
			o.Warn("Malformed Queue message talking to audience. Missing Score")
			sendQueueFailureResponse("Missing Score", enc)
			return
		}
		if nil == outobj.Scope {
			o.Warn("Malformed Queue message talking to audience. Missing Scope")
			sendQueueFailureResponse("Missing Scope", enc)
			return
		}
		if nil == outobj.Players || len(outobj.Players) < 1 {
			o.Warn("Malformed Queue message talking to audience. Missing Players")
			sendQueueFailureResponse("Missing Players", enc)
			return
		}
		for _, player := range outobj.Players {
			if !HostAuthorised(player) {
				o.Warn("Malformed Queue message - unknown player %s specified.", player)
				sendQueueFailureResponse("Invalid Player", enc)
				return
			}
		}
		job := NewRequest()
		job.Score = *outobj.Score
		switch (*outobj.Scope) {
		case "one":
			job.Scope = o.SCOPE_ONEOF
		case "all":
			job.Scope = o.SCOPE_ALLOF
		default:
			sendQueueFailureResponse("Invalid Scope", enc)
			return
		}
		job.Players = outobj.Players
		job.Params = outobj.Params

		QueueJob(job)
		sendQueueSuccessResponse(job, enc)
	default:
		o.Warn("Unknown operation talking to audience: \"%s\"", *(outobj.Op))
		return
	}

	_ = enc
}

func sendQueueSuccessResponse(job *o.JobRequest, enc *json.Encoder) {
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
}

func sendQueueFailureResponse(reason string, enc *json.Encoder) {
	resp := make([]interface{},2)
	resperr := new(string)
	*resperr = "Error"
	resp[0] = resperr
	if reason != "" {
		resp[1] = &reason
	}
	err := enc.Encode(resp)
	if nil != err {
		o.Warn("Couldn't encode response to audience: %s", err)
	}
}

func AudienceListener(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			o.Warn("Accept() failed on Audience Listenter.")
			break
		}
		go handleAudienceRequest(c)
	}
}

func UnixAudienceListener(sockaddr string) {
	fi, err := os.Stat(sockaddr)
	if err == nil {
		if fi.IsSocket() {
			o.Warn("Removing stale socket at %s", sockaddr)
			os.Remove(sockaddr)
		} else {
			o.Fail("%s exists and is not a socket", sockaddr)
		}
	}
	laddr, err := net.ResolveUnixAddr("unix", sockaddr)
	o.MightFail("Couldn't resolve audience socket address", err)
	l, err := net.ListenUnix("unix", laddr)
	o.MightFail("Couldn't start audience unixsock listener", err)
	// Fudge the permissions on the unixsock!
	fi, err = os.Stat(sockaddr)
	if err == nil {
		os.Chmod(sockaddr, fi.Mode | 0777)
	} else {
		o.Warn("Couldn't fudge permission on audience socket: %s", err)
	}
	
	// make sure we clean up the unix socket when we die.
	defer l.Close()
	defer os.Remove(sockaddr)
	AudienceListener(l)	
}

func StartAudienceSock() {
	go UnixAudienceListener(*AudienceSock)
}