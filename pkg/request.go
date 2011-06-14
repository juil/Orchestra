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
	JOB_ACCEPTED	= iota
	JOB_FINISHED
	JOB_FAILED

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
	Results		map[string]TaskResponse
}
type TaskRequest struct {
	Job		*JobRequest
	Player		string
	State		int
}
type TaskResponse struct {
	Response	map[string]string
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
	j = new(JobRequest)
	
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

func NewRequest() (req *JobRequest) {
	req = new(JobRequest)
	return req
}

func (req *JobRequest) normalise() {
	if (len(req.Players) > 1) {
		/* sort targets so search works */
		sort.SortStrings(req.Players)
	} else {
		if (req.Scope == SCOPE_ONEOF) {
			req.Scope = SCOPE_ALLOF
		}
	}
}

func (req *JobRequest) Tasks() (tasks []*TaskRequest) {
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
	return true
}
