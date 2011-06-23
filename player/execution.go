// execution.go

package main

import (
	"os"
	"bufio"
	o "orchestra"
)

func ExecuteJob(job *o.JobRequest) (complete chan int) {
	complete  = make(chan int, 1)
	go doExecution(job, complete)

	return complete
}

func batchLogger(jobid uint64, errpipe *os.File) {
	r := bufio.NewReader(errpipe)
	for {
		lb, _, err := r.ReadLine()
		if err == os.EOF {
			return
		}
		if err != nil {
			o.Warn("executionLogger failed: %s", err)
			return
		}
		o.Warn("JOB %d:%s", jobid, string(lb))
	}
}

func doExecution(job *o.JobRequest, completionChannel chan int) {
	// we must notify the parent when we exit.
	defer func(c chan int) { c <- 1 }(completionChannel)

	si := NewScoreInterface(job)
	if si == nil {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		return
	}
	score := Scores[job.Score]
	if !si.Prepare() {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		return
	}
	defer si.Cleanup()

	eenv := si.SetupProcess()
	job.MyResponse.State = o.RESP_RUNNING

	procenv := new(os.ProcAttr)
	for k, v := range eenv.Environment {
		procenv.Env = append(procenv.Env, k+"="+v)
	}

	procenv.Files = make([]*os.File, 3)
	lr, lw, err := os.Pipe()
	o.MightFail("Couldn't create pipe", err)
	procenv.Files[2] = lw
	if nil != eenv.Files {
		for i := range eenv.Files {
			if i < 2 {
				procenv.Files[i] = eenv.Files[i]
			} else {
				procenv.Files = append(procenv.Files, eenv.Files[i])
			}
		}
	}	
	var args []string = nil
	args = append(args, eenv.Arguments...)

	go batchLogger(job.Id, lr)
	proc, err := os.StartProcess(score.Executable, args, procenv)
	if err != nil {
		job.MyResponse.State = o.RESP_FAILED_HOST_ERROR
		return
	}
	wm, err := proc.Wait(0)
	if err != nil {
		job.MyResponse.State = o.RESP_FAILED_UNKNOWN
		// Worse of all, we don't even know if we succeeded.
		return
	}
	if !(wm.WaitStatus.Signaled() || wm.WaitStatus.Exited()) {
		o.Assert("Non Terminal notification received when not expected.")
		return
	}
	if wm.WaitStatus.Signaled() {
		job.MyResponse.State = o.RESP_FAILED_UNKNOWN
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