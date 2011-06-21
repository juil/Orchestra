// registry.go
//
// Job Registry.
//
// The Registry provides a 'threadsafe' interface to various global
// information stores.
//
// The registry dispatch thread is forbidden from performing any work
// that is likely to block.  Result channels must be buffered with
// enough space for the full set of results.
//
// This registry implementation implements stores that will be needed
// on both the conductor and the player.

package orchestra

import (
	"sort"
)

const (
	requestAddJob			= iota
	requestGetJob
	requestAddJobResult
	requestGetJobResult
	requestGetJobResultNames
	requestDisqualifyPlayer
	requestReviewJobStatus

	requestQueueSize		= 10
)


type registryRequest struct {
	operation		int
	id			uint64
	player			string
	job			*JobRequest
	tresp			*TaskResponse
	responseChannel		chan *registryResponse
}

type registryResponse struct {
	success			bool
	job			*JobRequest
	tresp			*TaskResponse
	names			[]string
}
	
var chanRequest = make(chan *registryRequest, requestQueueSize)

// bake a minimal request structure together.
func newRequest(wants_response bool) (r *registryRequest) {
	r = new(registryRequest)
	if wants_response {
		r.responseChannel = make(chan *registryResponse, 1)
	}

	return r
}

// Add a Job to the registry.  Return true if successful, returns
// false if the job is lacking critical information (such as a JobId)
// and can't be registered.
func JobAdd(job *JobRequest) bool {
	rr := newRequest(true)
	rr.operation = requestAddJob
	rr.job = job

	chanRequest <- rr
	resp := <- rr.responseChannel 
	return resp.success
}

// Get a Job from the registry.  Returns the job if successful,
// returns nil if the job couldn't be found.
func JobGet(id uint64) *JobRequest {
	rr := newRequest(true)
	rr.operation = requestGetJob
	rr.id = id

	chanRequest <- rr
	resp := <- rr.responseChannel
	return resp.job
}

// Attach a result to a Job in the Registry
//
// This exists in order to prevent nasty concurrency problems
// when trying to put results back onto the job.  Reading a job is far
// less of a problem than writing to it.
func JobAddResult(playername string, task *TaskResponse) bool {
	rr := newRequest(true)
	rr.operation = requestAddJobResult
	rr.tresp = task
	rr.player = playername
	chanRequest <- rr
	resp := <- rr.responseChannel
	return resp.success
}

// Get a result from the registry
func JobGetResult(id uint64, playername string) (tresp *TaskResponse) {
	rr := newRequest(true)
	rr.operation = requestGetJobResult
	rr.id = id
	rr.player = playername
	chanRequest <- rr
	resp := <- rr.responseChannel
	return resp.tresp
}

// Get a list of names we have results for against a given job.
func JobGetResultNames(id uint64) (names []string) {
	rr := newRequest(true)
	rr.operation = requestGetJobResultNames
	rr.id = id

	chanRequest <- rr
	resp := <- rr.responseChannel 
	return resp.names
}

//  Disqualify a player from servicing a job
func JobDisqualifyPlayer(id uint64, playername string) bool {
	rr := newRequest(true)
	rr.operation = requestDisqualifyPlayer
	rr.id = id
	rr.player = playername

	chanRequest <- rr
	resp := <- rr.responseChannel

	return resp.success
}

func JobReviewState(id uint64) bool {
	rr := newRequest(true)
	rr.operation = requestReviewJobStatus
	rr.id = id

	chanRequest <- rr
	resp := <- rr.responseChannel

	return resp.success
}

func manageRegistry() {
	jobRegister := make(map[uint64]*JobRequest)

	for {
		req := <- chanRequest
		resp := new (registryResponse)
		switch (req.operation) {
		case requestAddJob:
			if nil != req.job {
				// ensure that the players are sorted!
				sort.SortStrings(req.job.Players)
				jobRegister[req.job.Id] = req.job
				resp.success = true
			} else {
				resp.success = false
			}
		case requestGetJob:
			job, exists := jobRegister[req.id]
			resp.success = exists
			if exists {
				resp.job = job
			}
		case requestAddJobResult:
			job, exists := jobRegister[req.tresp.Id]
			resp.success = exists
			if exists {
				job.results[req.player] = req.tresp
			}
		case requestGetJobResult:
			job, exists := jobRegister[req.id]
			if exists {
				result, exists := job.results[req.player]
				resp.success = exists
				if exists {
					resp.tresp = result
				}
			} else {
				resp.success = false
			}
		case requestGetJobResultNames:
			job, exists := jobRegister[req.id]
			resp.success = exists
			if exists {
				resp.names = make([]string, len(job.results))
				idx := 0
				for k, _ := range job.results {
					resp.names[idx] = k
					idx++
				}
			}
		case requestDisqualifyPlayer:
			job, exists := jobRegister[req.id]
			if exists {
				idx := sort.Search(len(job.Players), func(idx int) bool { return job.Players[idx] >= req.player })
				if (job.Players[idx] == req.player) {
					resp.success = true
					newplayers := make([]string, len(job.Players)-1)
					copy(newplayers[0:idx], job.Players[0:idx])
					copy(newplayers[idx:len(job.Players)-1], job.Players[idx+1:len(job.Players)])
					job.Players = newplayers
				} else {
					resp.success = false
				}
			} else {
				resp.success = false
			}
		case requestReviewJobStatus:
			job, exists := jobRegister[req.id]
			resp.success = exists
			if exists {
				switch job.Scope {
				case SCOPE_ONEOF:
					// look for a success (any success) in the responses
					var success bool = false
					for _, res := range job.results {
						if res.State == RESP_FINISHED {
							success = true
							break
						}
					}
					// update the job state based upon these findings
					if success {
						job.State = JOB_SUCCESSFUL
					} else {
						if len(job.Players) < 1 {
							job.State = JOB_FAILED
						} else {
							job.State = JOB_PENDING
						}
					}
				case SCOPE_ALLOF:
					var success int = 0
					var failed  int = 0

					for pidx := range job.Players {
						p := job.Players[pidx]
						resp, exists := job.results[p]
						if exists {
							if resp.DidFail() {
								failed++
							} else if resp.State == RESP_FINISHED {
								success++
							}
						}
					}
					if (success + failed) < len(job.Players) {
						job.State = JOB_PENDING
					} else if success == len(job.Players) {
						job.State = JOB_SUCCESSFUL
					} else if failed == len(job.Players) {
						job.State = JOB_FAILED
					} else {
						job.State = JOB_FAILED_PARTIAL
					}
				}
			}
		}
		if req.responseChannel != nil {
			req.responseChannel <- resp
		}
	}
}

func init() {
	go manageRegistry()
}
