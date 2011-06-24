/* player.go
*/

package main

import (
	"os"
	"fmt"
	"flag"
	o	"orchestra"
	"crypto/tls"
	"net"
	"time"
	"container/list"
)

const (
	InitialReconnectDelay		= 5e9
	MaximumReconnectDelay		= 300e9
	ReconnectDelayScale		= 2
	KeepaliveDelay 			= 30e9
	RetryDelay			= 5e9
)

var (
	masterHostname = flag.String("master", "conductor", "Hostname of the Conductor")
	localHostname  = flag.String("hostname", "", "My Hostname (defaults to autoprobing)")
	masterPort     = flag.Int("port", o.DefaultMasterPort, "Port to contact conductor on")
	ScoreDirectory = flag.String("score-dir", "/usr/lib/orchestra/scores", "Path for the Directory containing valid scores")
 	receivedMessage = make(chan *o.WirePkt)
	lostConnection = make(chan int)
	pendingQueue	= list.New()
	unacknowledgedQueue	= list.New()
	newConnection	= make(chan net.Conn)
)

func getNextPendingJob() (job *o.JobRequest) {
	e := pendingQueue.Front()
	if e != nil {
		job, _ = e.Value.(*o.JobRequest)
		pendingQueue.Remove(e)
	}
	return job
}

func appendPendingJob(job *o.JobRequest) {
	pendingQueue.PushBack(job)
}

func getNextUnacknowledgedResponse() (resp *o.TaskResponse) {
	e := unacknowledgedQueue.Front()
	if e != nil {
		resp, _ = e.Value.(*o.TaskResponse)
		unacknowledgedQueue.Remove(e)
	}
	return resp
}

func appendUnacknowledgedResponse(resp *o.TaskResponse) {
	resp.RetryTime = time.Nanoseconds() + RetryDelay
	unacknowledgedQueue.PushBack(resp)
}

func acknowledgeResponse(jobid uint64) {
	for e := unacknowledgedQueue.Front(); e != nil; e = e.Next() {
		resp := e.Value.(*o.TaskResponse)
		if resp.Id == jobid {
			unacknowledgedQueue.Remove(e)
		}
	}
}

func sendResponse(c net.Conn, resp *o.TaskResponse) {
	//FIXME: update retry time on Response
	o.Warn("Sending Response!")
	ptr := resp.Encode()
	p, err := o.Encode(ptr)
	o.MightFail("Failed to encode response", err)
	_, err = p.Send(c)
	if err != nil {
		o.Warn("Transmission error: %s", err)
		c.Close()
		prequeueResponse(resp)
		lostConnection <- 1
	} else {
		appendUnacknowledgedResponse(resp)
	}
}

func prequeueResponse(resp *o.TaskResponse) {
	unacknowledgedQueue.PushFront(resp)
}

func Reader(conn net.Conn) {
	defer func(l chan int) {
		l <- 1
	}(lostConnection)

	for {
		pkt, err := o.Receive(conn)
		if (err != nil) {
			o.Warn("Error receiving message: %s", err)
			break;
		}
		receivedMessage <- pkt
	}	
}

func handleNop(c net.Conn, message interface{}) {
	o.Warn("NOP Received")
}

func handleIllegal(c net.Conn, message interface{}) {
	o.Fail("Got Illegal Message")
}

func handleRequest(c net.Conn, message interface{}) {
	o.Warn("Request Recieved.  Decoding!")
	ptr, ok := message.(*o.ProtoTaskRequest)
	if !ok {
		o.Assert("CC stuffed up - handleRequest got something that wasn't a ProtoTaskRequest.")
	}
	job := o.JobFromProto(ptr)
	/* search the registry for the job */
	o.Warn("Request for Job.ID %d", job.Id)
	existing := o.JobGet(job.Id)
	if nil != existing {
		if (existing.MyResponse.IsFinished()) {
			o.Warn("job%d: Resending Response", job.Id)
			sendResponse(c, existing.MyResponse)
		}
	} else {
		// check to see if we have the score
		// add the Job to our Registry
		job.MyResponse = o.NewTaskResponse()
		job.MyResponse.State = o.RESP_PENDING		
		o.JobAdd(job)
		o.Warn("Added Job %d to our local registry", job.Id)
		// and then push it onto the pending job list so we know it needs actioning.
		appendPendingJob(job)
	}
}

func handleAck(c net.Conn, message interface{}) {
	o.Warn("Ack Received")
	ack, ok := message.(*o.ProtoAcknowledgement)
	if !ok {
		o.Assert("CC stuffed up - handleAck got something that wasn't a ProtoAcknowledgement.")
	}
	if ack.Id != nil {
		acknowledgeResponse(*ack.Id)
	}
}


var dispatcher	= map[uint8] func(net.Conn, interface{}) {
	o.TypeNop:		handleNop,
	o.TypeTaskRequest:	handleRequest,
	o.TypeAcknowledgement:	handleAck,

	/* P->C only messages, should never appear on the wire to us. */
	o.TypeIdentifyClient:	handleIllegal,
	o.TypeReadyForTask:	handleIllegal,
	o.TypeTaskResponse:	handleIllegal,
}



func connectMe() {
	var backOff int64 = InitialReconnectDelay
	for {
		tconf := &tls.Config{
		}
		raddr := fmt.Sprintf("%s:%d", *masterHostname, *masterPort)
		o.Warn("Connecting to %s", raddr)
		conn, err := tls.Dial("tcp", raddr, tconf)
		
		if err != nil {
			o.Warn("Couldn't connect to master: %s", err)
			o.Warn("Sleeping for %d seconds", backOff/1e9)
			err := time.Sleep(backOff)
			o.MightFail("Couldn't Sleep",err)

			backOff *= ReconnectDelayScale
			if backOff > MaximumReconnectDelay {
				backOff = MaximumReconnectDelay
			}
		} else {
			newConnection <- conn
			return
		}
	}
}


func ProcessingLoop() {
	var	conn			net.Conn		= nil
	var     nextRetryResp		*o.TaskResponse 	= nil
	var	jobCompletionChan	<-chan *o.TaskResponse	= nil
	var	pendingTaskRequest	bool			= false
	// kick off a new connection attempt.
	go connectMe()

	// and this is where we spin!
	for {	
		var retryDelay int64 = 0
		var retryChan  <-chan int64 = nil

		if conn != nil {
			for nextRetryResp == nil {
				nextRetryResp = getNextUnacknowledgedResponse()
				if nil == nextRetryResp {
					break
				}
				retryDelay = nextRetryResp.RetryTime - time.Nanoseconds()
				if retryDelay < 0 {
					sendResponse(conn, nextRetryResp)
					nextRetryResp = nil
				}
			}
			if nextRetryResp != nil {
				retryChan = time.After(retryDelay)
			}
		}
		if jobCompletionChan == nil {
			nextJob := getNextPendingJob()
			if nextJob != nil {
				jobCompletionChan = ExecuteJob(nextJob)
			} else {
				if conn != nil && !pendingTaskRequest {
					o.Warn("Asking for trouble")
					p := o.MakeReadyForTask()
					p.Send(conn)
					o.Warn("Sent Request for trouble")
					pendingTaskRequest = true
				}
			}
		}
		select {
		// Currently executing job finishes.
		case newresp := <- jobCompletionChan:
			// preemptively set a retrytime.
			newresp.RetryTime = time.Nanoseconds()
			// ENOCONN - sub it in as our next retryresponse, and prepend the old one onto the queue.
			if nil == conn {
				if nil != nextRetryResp {
					prequeueResponse(nextRetryResp)
				}
				o.Warn("job%d: Queuing Initial Response", newresp.Id)
				nextRetryResp = newresp
			} else {
				o.Warn("job%d: Sending Initial Response", newresp.Id)
				sendResponse(conn, newresp)
			}
			jobCompletionChan = nil
		// If the current unacknowledged response needs a retry, send it.
		case <-retryChan:
			sendResponse(conn, nextRetryResp)
			nextRetryResp = nil
		// New connection.  Set up the receiver thread and Introduce ourselves.
		case newconn := <-newConnection:
			if conn != nil {
				conn.Close()
			}
			conn = newconn

			// start the reader
			go Reader(conn)
		
			/* Introduce ourself */
			p := o.MakeIdentifyClient(*localHostname)
			p.Send(conn)
		// Lost connection.  Shut downt he connection.
		case <-lostConnection:
			o.Warn("Lost Connection to Master")
			conn.Close()
			conn = nil
			// restart the connection attempts
			go connectMe()
		// Message received from master.  Decode and action.
		case p := <-receivedMessage:
			// because the message could possibly be an ACK, push the next retry response back into the queue so acknowledge can find it.
			if nil != nextRetryResp {
				prequeueResponse(nextRetryResp)
				nextRetryResp = nil
			}
			var upkt interface{} = nil
			if p.Length > 0 {
				var err os.Error
				upkt, err = p.Decode()
				o.MightFail("Couldn't decode packet from master", err)
			}
			handler, exists := dispatcher[p.Type]
			if (exists) {
				handler(conn, upkt)
			} else {
				o.Fail("Unhandled Pkt Type %d", p.Type)
			}
		// Keepalive delay expired.  Send Nop.
		case <-time.After(KeepaliveDelay):
			if conn == nil {
				break
			}
			o.Warn("Sending Nop")
			p := o.MakeNop()
			p.Send(conn)
		}
	}
}

func main() {
	flag.Parse()

	if (*localHostname == "") {
		*localHostname = o.ProbeHostname()
		o.Warn("No hostname provided - probed hostname: %s", *localHostname)
	}
	LoadScores()
	ProcessingLoop()
}
