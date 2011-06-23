// if_env
//
// 'env' score interface

package main

import (
	o "orchestra"
)

const (
	EnvironmentPrefix = "ORC_"
)

func init() {
	RegisterInterface("env", newEnvInterface)
}

type EnvInterface struct {
	job	*o.JobRequest
}

func newEnvInterface(job *o.JobRequest) (iface ScoreInterface) {
	ei := new(EnvInterface)
	ei.job = job

	return ei
}

func (ei *EnvInterface) Prepare() bool {
	// does nothing!
	return true
}

func (ei *EnvInterface) SetupProcess() (ee *ExecutionEnvironment) {
	ee = NewExecutionEnvironment()
	for k,v := range ei.job.Params {
		ee.Environment[EnvironmentPrefix+k] = v
	}

	return ee
}

func (ei *EnvInterface) Cleanup() {
	// does nothing!
}