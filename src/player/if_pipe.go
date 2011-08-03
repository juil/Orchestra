// if_pipe
//
// 'pipe' score interface
//
// The PIPE score interface works much like the ENV interace, but also attaches
// a pipe to STDOUT which captures <x>=<y> repsonse values.

package main

import (
	"strings"
	"bufio"
	"os"
	o "orchestra"
)

const (
	pipeEnvironmentPrefix = "ORC_"
)

func init() {
	RegisterInterface("pipe", newPipeInterface)
}

type PipeInterface struct {
	job	*o.JobRequest
	pipew	*os.File
}

func newPipeInterface(job *o.JobRequest) (iface ScoreInterface) {
	ei := new(PipeInterface)
	ei.job = job

	return ei
}


// pipeListener is the goroutine that sits on the stdout pipe and
// processes what it sees.
func pipeListener(job *o.JobRequest, outpipe *os.File) {
	defer outpipe.Close()

	r := bufio.NewReader(outpipe)
	for {
		lb, _, err := r.ReadLine()
		if err == os.EOF {
			return
		}
		if err != nil {
			o.Warn("pipeListener failed: %s", err)
			return
		}
		linein := string(lb)
		if strings.Index(linein, "=") >= 0 {
			bits := strings.SplitN(linein, "=", 2)
			job.MyResponse.Response[bits[0]] = bits[1]
		}
	}
}


func (ei *PipeInterface) Prepare() bool {
	lr, lw, err := os.Pipe()
	if (err != nil) {
		return false
	}
	// save the writing end of the pipe so we can close our local end of it during cleanup.
	ei.pipew = lw

	// start the pipe listener
	go pipeListener(ei.job, lr)

	return true
}

func (ei *PipeInterface) SetupProcess() (ee *ExecutionEnvironment) {
	ee = NewExecutionEnvironment()
	for k,v := range ei.job.Params {
		ee.Environment[pipeEnvironmentPrefix+k] = v
	}
	ee.Files = make([]*os.File, 2)
	ee.Files[1] = ei.pipew
	return ee
}

func (ei *PipeInterface) Cleanup() {
	// close the local copy of the pipe.
	//
	// if the child didn't start, this will also EOF the
	// pipeListener which will clean up that goroutine.
	ei.pipew.Close()
}