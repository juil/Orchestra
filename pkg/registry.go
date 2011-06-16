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

const (
	requestAddJob			= iota
	requestGetJob
	requestAddJobResult
	requestGetJobResult
	requestGetJobResultNames

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

func manageRegistry() {
	jobRegister := make(map[uint64]*JobRequest)

	for {
		req := <- chanRequest
		resp := new (registryResponse)
		switch (req.operation) {
		case requestAddJob:
			if nil != req.job {
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
		}
		if req.responseChannel != nil {
			req.responseChannel <- resp
		}
	}
}

func init() {
	go manageRegistry()
}