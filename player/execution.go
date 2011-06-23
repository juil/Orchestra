// execution.go

package main

import (
	"os"
	o "orchestra"
)

func ExecuteJob(job *o.JobRequest) (complete chan int) {
	complete  = make(chan int, 1)
	go doExecution(job, complete)

	return complete
}

func doExecution(job *o.JobRequest, completionChannel chan int) {
	// we must notify the parent when we exit.
	defer func(c chan int) { c <- 1 }(completionChannel)

	si := NewScoreInterface(job)
	if si == nil {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		//FIXME: error the job properly.
		return
	}
	score := Scores[job.Score]
	if !si.Prepare() {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		//FIXME: error the job properly.
		return
	}
	defer si.Cleanup()

	eenv := si.SetupProcess()
	job.MyResponse.State = o.RESP_RUNNING

	procenv := new(os.ProcAttr)
	for k, v := range eenv.Environment {
		procenv.Env = append(procenv.Env, k+"="+v)
	}
	var args []string = nil
	args = append(args, eenv.Arguments...)

	proc, err := os.StartProcess(score.Executable, args, procenv)
	if err != nil {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		//FIXME: error the job properly
		return
	}
	wm, err := proc.Wait(0)
	if err != nil {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		//FIXME: error the job properly
		// Worse of all, we don't even know if we succeeded.
		return
	}
	if !(wm.WaitStatus.Signaled() || wm.WaitStatus.Exited()) {
		o.Assert("Non Terminal notification received when not expected.")
		return
	}
	if wm.WaitStatus.Signaled() {
		job.MyResponse.State = o.RESP_FAILED
		return
	}
	if wm.WaitStatus.Exited() {
		if 0 == wm.WaitStatus.ExitStatus() {
			job.MyResponse.State = o.RESP_FINISHED
		} else {
			job.MyResponse.State = o.RESP_FAILED
		}
		return
	}
	o.Assert("Should never get here.")
}