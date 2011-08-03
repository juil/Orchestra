/* request.go
 *
 * Request objects, half the marshalling code. 
*/

package orchestra

import (
	"sort"
	"goprotobuf.googlecode.com/hg/proto"
)

const (
	// Task is fresh and has never been sent to the client.  It can be rescheduled still.
	TASK_QUEUED	= iota
	// Task has been acknowledged by host and is executing.
	TASK_PENDINGRESULT
	// Task has finished and we have received a result.
	TASK_FINISHED

	// Job is pending resolution
	JOB_PENDING
	// Job has completed and has no failures.
	JOB_SUCCESSFUL
	// Job has completed and has mixed results.
	JOB_FAILED_PARTIAL
	// Job has completed and has completely failed.
	JOB_FAILED

	// Response states
	RESP_PENDING	// internal state, not wire.
	RESP_RUNNING
	RESP_FINISHED
	RESP_FAILED
	RESP_FAILED_UNKNOWN_SCORE
	RESP_FAILED_HOST_ERROR
	RESP_FAILED_UNKNOWN // unknown error.  it just didnt work.

	SCOPE_ONEOF
	SCOPE_ALLOF
)



type JobRequest struct {
	Score		string
	Scope		int
	Players		[]string
	Id		uint64
	State		int
	Params		map[string]string
	Tasks		[]*TaskRequest
	// These are private - you need to use the registry to access these
	results		map[string]*TaskResponse

	// these fields are used by the player only
	MyResponse	*TaskResponse
	RetryTime	int64
}
type TaskRequest struct {
	Job		*JobRequest
	Player		string
	State		int
	RetryTime	int64
}
type TaskResponse struct {
	State		int
	Id		uint64
	Response	map[string]string
	// player only fields
	RetryTime	int64
}


/* ugh ugh ugh.  As much as I love protocol buffers, not having maps
 * as a native type is a PAIN IN THE ASS.
 *
 * Here's some common code to convert my K/V format in protocol
 * buffers to and from native Go structures.
*/

func mapFromJobParameters(parray []*ProtoJobParameter) (mapparam map[string]string) {
	mapparam = make(map[string]string)

	for p := range parray {
		mapparam[*(parray[p].Key)] = *(parray[p].Value)
	}

	return mapparam
}

func jobParametersFromMap(mapparam map[string]string) (parray []*ProtoJobParameter) {
	parray = make([]*ProtoJobParameter, len(mapparam))
	i := 0
	for k,v := range mapparam {
		arg := new(ProtoJobParameter)
		arg.Key = proto.String(k)
		arg.Value = proto.String(v)
		parray[i] = arg
		i++
	}

	return parray
}

/* oh noes!  assymetry! the wire tasks map better to Jobs, but we never send a job without 
 * a task structure
*/
func JobFromProto(ptr *ProtoTaskRequest) (j *JobRequest) {
	j = NewJobRequest()
	
	j.Score = *(ptr.Jobname)
	j.Id = *(ptr.Id)
	j.Params = mapFromJobParameters(ptr.Parameters)

	return j
}

func (task *TaskRequest) Encode() (ptr *ProtoTaskRequest) {
	ptr = new(ProtoTaskRequest)
	ptr.Jobname = &task.Job.Score
	ptr.Id = new(uint64)
	*ptr.Id = task.Job.Id
	ptr.Parameters = jobParametersFromMap(task.Job.Params)

	return ptr
}

func NewJobRequest() (req *JobRequest) {
	req = new(JobRequest)
	req.results = make(map[string]*TaskResponse)
	return req
}

func (req *JobRequest) normalise() {
	if (len(req.Players) > 1) {
		/* sort targets so search works */
		sort.Strings(req.Players)
	} else {
		if (req.Scope == SCOPE_ONEOF) {
			req.Scope = SCOPE_ALLOF
		}
	}
}

func (req *JobRequest) MakeTasks() (tasks []*TaskRequest) {
	req.normalise()

	var numtasks int
	
	switch (req.Scope) {
	case SCOPE_ONEOF:
		numtasks = 1
	case SCOPE_ALLOF:
		numtasks = len(req.Players)
	}
	tasks = make([]*TaskRequest, numtasks)
	
	for c := 0; c < numtasks; c++ {
		t := new(TaskRequest)
		t.State = TASK_QUEUED
		t.Job = req
		if (req.Scope == SCOPE_ALLOF) {
			t.Player = req.Players[c]
		}
		tasks[c] = t
	}
	return tasks
}

func (req *JobRequest) Valid() bool {
	if (len(req.Players) <= 0) {
		return false
	}
	return true
}


/** Task Related Magic */

func (task *TaskRequest) IsTarget(player string) (valid bool) {
	valid = false
	if task.Player == "" {
		n := sort.SearchStrings(task.Job.Players, player)
		if task.Job.Players[n] == player {
			valid = true
		}
	} else {
		if task.Player == player {
			valid = true
		}
	}
	return valid
}


// Response related magic

func NewTaskResponse() (resp *TaskResponse) {
	resp = new(TaskResponse)
	resp.Response = make(map[string]string)

	return resp
}

func (resp *TaskResponse) Encode() (ptr *ProtoTaskResponse) {
	ptr = new(ProtoTaskResponse)
	
	switch resp.State {
	case RESP_RUNNING:
		ptr.Status = NewProtoTaskResponse_TaskStatus(ProtoTaskResponse_JOB_INPROGRESS)
	case RESP_FINISHED:
		ptr.Status = NewProtoTaskResponse_TaskStatus(ProtoTaskResponse_JOB_SUCCESS)
	case RESP_FAILED:
		ptr.Status = NewProtoTaskResponse_TaskStatus(ProtoTaskResponse_JOB_FAILED)
	case RESP_FAILED_UNKNOWN_SCORE:
		ptr.Status = NewProtoTaskResponse_TaskStatus(ProtoTaskResponse_JOB_UNKNOWN)
	case RESP_FAILED_HOST_ERROR:
		ptr.Status = NewProtoTaskResponse_TaskStatus(ProtoTaskResponse_JOB_HOST_FAILURE)
	case RESP_FAILED_UNKNOWN:
		ptr.Status = NewProtoTaskResponse_TaskStatus(ProtoTaskResponse_JOB_UNKNOWN_FAILURE)
	}
	ptr.Id = new(uint64)
	*ptr.Id = resp.Id
	ptr.Response = jobParametersFromMap(resp.Response)

	return ptr
}


func (resp *TaskResponse) IsFinished() bool {
	switch resp.State {
	case RESP_FINISHED:
		fallthrough
	case RESP_FAILED:
		fallthrough
	case RESP_FAILED_UNKNOWN_SCORE:
		fallthrough
	case RESP_FAILED_HOST_ERROR:
		fallthrough
	case RESP_FAILED_UNKNOWN:
		return true
	}
	return false
}

// true if the task failed.  false otherwise.
func (resp *TaskResponse) DidFail() bool {
	// we actually test for the known successes or in progress
	// states.  Everything else must be failure.
	switch resp.State {
	case RESP_RUNNING:
		fallthrough
	case RESP_FINISHED:
		return false
	}
	return true
}

// true if the task can be tried.
// precond:  DidFail is true, job is a ONE_OF job.
// must return false otherwise.
func (resp *TaskResponse) CanRetry() bool {
	switch resp.State {
	case RESP_FAILED_UNKNOWN_SCORE:
		fallthrough
	case RESP_FAILED_HOST_ERROR:
		return true
	}
	return false
}


func ResponseFromProto(ptr *ProtoTaskResponse) (r *TaskResponse) {
	r = new(TaskResponse)

	switch (*(ptr.Status)) {
	case ProtoTaskResponse_JOB_INPROGRESS:
		r.State = RESP_RUNNING
	case ProtoTaskResponse_JOB_SUCCESS:
		r.State = RESP_FINISHED
	case ProtoTaskResponse_JOB_FAILED:
		r.State = RESP_FAILED
	case ProtoTaskResponse_JOB_HOST_FAILURE:
		r.State = RESP_FAILED_HOST_ERROR
	case ProtoTaskResponse_JOB_UNKNOWN:
		r.State = RESP_FAILED_UNKNOWN_SCORE
	case ProtoTaskResponse_JOB_UNKNOWN_FAILURE:
		fallthrough
	default:
		r.State = RESP_FAILED_UNKNOWN
	}

	r.Id = *(ptr.Id)
	r.Response = mapFromJobParameters(ptr.Response)

	return r
}
