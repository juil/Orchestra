/* dispatch.go
*/

package main;

import (
	"sync/atomic"
	"container/list"
	o "orchestra"
)

var lastId uint64 = 0

func nextRequestId() uint64 {
	return atomic.AddUint64(&lastId, 1)
}

func NewRequest() (req *o.JobRequest) {
	req = o.NewJobRequest()

	return req
}

const messageBuffer = 10

var newJob		= make(chan *o.JobRequest, messageBuffer)
var rqTask		= make(chan *o.TaskRequest, messageBuffer)
var playerIdle		= make(chan *ClientInfo, messageBuffer)
var playerDead		= make(chan *ClientInfo, messageBuffer)
var statusRequest	= make(chan(chan *QueueInformation))

func PlayerWaitingForJob(player *ClientInfo) {
	playerIdle <- player
}

func PlayerDied(player *ClientInfo) {
	playerDead <- player
}

func DispatchTask(task *o.TaskRequest) {
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

func QueueJob(job *o.JobRequest) {
	/* first, allocate the Job it's ID */
	job.Id = nextRequestId()
	/* first up, split the job up into it's tasks. */
	job.Tasks = job.MakeTasks()
	/* add it to the registry */
	o.JobAdd(job)
	/* an enqueue all of the tasks */
	for i := range job.Tasks {
		DispatchTask(job.Tasks[i])
	}
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
				t,_ := i.Value.(*o.TaskRequest)
				if t.IsTarget(player.Player) {
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
				if player.Player == p.Player {
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
				if task.IsTarget(p.Player) {
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
				response.idlePlayers[idx] = player.Player
				idx++
			}
			respChan <- response
		}
	}
}
