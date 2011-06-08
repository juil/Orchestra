/* dispatch.go
*/

package main;

import (
	"sort"
	"sync/atomic"
	"container/list"
	o "orchestra"
)

var lastId uint64 = 0

func nextRequestId() uint64 {
	return atomic.AddUint64(&lastId, 1)
}

const (
	OneOf = iota
	AllOf

	Queued
	InProgress
	Completed
	JobFailed
	HostFailed
	JobUnknown
	UnknownFailure
)

type JobResult struct {
	Id		uint64
	Success		uint64
	Response	map[string] string
	
}

/* this is the actual 'task' to execute
 * which gets handed off elsewhere.
 */
type JobTask struct {
	job	*JobRequest
	player	string
}

func (req *JobRequest) MakeTasks() (tasks []*JobTask) {
	var numtasks int
	
	switch (req.Scope) {
	case OneOf:
		numtasks = 1
	case AllOf:
		numtasks = len(req.Players)
	}
	tasks = make([]*JobTask, numtasks)
	
	for c := 0; c < numtasks; c++ {
		t := new(JobTask)
		t.job = req
		if (req.Scope == AllOf) {
			t.player = req.Players[c]
		}
		tasks[c] = t
	}
	return tasks
}

type JobRequest struct {
	Id		uint64
	Scope		int
	Players		[]string
	Params		map[string] string
	Results		map[string] JobResult
}

func NewRequest() (req *JobRequest) {
	req = new(JobRequest)
	
	req.Id = nextRequestId()

	return req
}

/* normalise a request prepares it for execution.
 *
 * There's a few semantics we fix on the way.
 *
 * OneOf jobs for a single host get reduced to a AllOf job.
*/
func (req *JobRequest) cook() {
	if (len(req.Players) > 1) {
		/* sort targets so search works */
		sort.SortStrings(req.Players)
	} else {
		if (req.Scope == OneOf) {
			req.Scope = AllOf
		}
	}
}

func (req *JobRequest) valid() bool {
	if (len(req.Players) <= 0) {
		return false
	}

	return true
}

const messageBuffer = 10

var newJob		= make(chan *JobRequest, messageBuffer)
var rqTask		= make(chan *JobTask, messageBuffer)
var playerIdle		= make(chan *ClientInfo, messageBuffer)
var playerDead		= make(chan *ClientInfo, messageBuffer)
var statusRequest	= make(chan(chan *QueueInformation))

func (task *JobTask) IsTarget(player string) (valid bool) {
	valid = false
	if task.player == "" {
		n := sort.SearchStrings(task.job.Players, player)
		if task.job.Players[n] == player {
			valid = true
		}
	} else {
		if task.player == player {
			valid = true
		}
	}
	return true
}

func PlayerWaitingForJob(player *ClientInfo) {
	playerIdle <- player
}

func PlayerDied(player *ClientInfo) {
	playerDead <- player
}

func (task *JobTask) Dispatch() {
	rqTask <- task
}

type QueueInformation struct {
	idlePlayers 	[]string
	waitingTasks	int
}

func DispatchStatus() (waitingTasks int, waitingPlayers []string) {
	r := make(chan *QueueInformation)

	statusRequest <- r
	s := <- r

	return s.waitingTasks, s.idlePlayers
}

func InitDispatch() {
	go masterDispatch(); // go!
}

func masterDispatch() {
	pq := list.New()
	tq := list.New()

	for {
		select {
		case player := <-playerIdle:
			o.Warn("Dispatch: Player")
			/* first, scan to see if we have anything for this player */
			i := tq.Front()
			for {
				if (nil == i) {
					/* Out of items! */
					/* Append this player to the waiting players queue */
					pq.PushBack(player)
					break;
				}
				t,_ := i.Value.(*JobTask)
				if t.IsTarget(player.Hostname) {
					/* Found a valid job. Send it to the player, and remove it from our pending 
					 * list */
					tq.Remove(i)
					player.TaskQ <- t
					break;
				}
				i = i.Next()
			}
		case player := <-playerDead:
			o.Warn("Dispatch: Dead Player")
			for i := pq.Front(); i != nil; i = i.Next() {
				p, _ := i.Value.(*ClientInfo)
				if player.Hostname == p.Hostname {
					pq.Remove(i)
					break;
				}
			}
		case task := <-rqTask:
			o.Warn("Dispatch: Task")
			/* first, scan to see if we have valid pending player for this task */
			i := pq.Front()
			for {
				if (nil == i) {
					/* Out of players! */
					/* Append this task to the waiting tasks queue */
					tq.PushBack(task)
					break;
				}
				p,_ := i.Value.(*ClientInfo)
				if task.IsTarget(p.Hostname) {
					/* Found it. */
					pq.Remove(i)
					p.TaskQ <- task
					break;
				}
				i = i.Next();
			}
		case respChan := <-statusRequest:
			o.Warn("Status!")
			response := new(QueueInformation)
			response.waitingTasks = tq.Len()
			pqLen := pq.Len()
			response.idlePlayers = make([]string, pqLen)
			
			idx := 0
			for i := pq.Front(); i != nil; i = i.Next() {
				player,_ := i.Value.(*ClientInfo)
				response.idlePlayers[idx] = player.Hostname
				idx++
			}
			respChan <- response
		}
	}
}
