// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package progress

import (
	"fmt"
	"time"
)

// Client is the interface a frontend must implement in order to be notified about
// the install progress
type Client interface {
	// Desc is called when a new progress unit is started
	Desc(desc string)

	// Partial is called on behalf of a MultiStep progress task - it's called for each
	// partial step completion
	Partial(total int, step int)

	// Step is called on behalf of a Loop progress task - it's called based on the
	// time returned by LoopWaitDuration
	Step()

	// Done is called whenever a progress task is completed
	Done()

	// LoopWaitDuration gives the implementation the opportunity configure the loop progress
	// step period
	LoopWaitDuration() time.Duration
}

// Progress is the internal interface for the progress subsystem, currently we have
// two different implementations a MultiStep and a Loop based one
type Progress interface {
	Partial(step int)
	Done()
}

// BaseProgress is the common implementation between MultiStep and Loop progress
type BaseProgress struct {
	total int
}

// Loop defines the specific data for Loop progress implementation
type Loop struct {
	BaseProgress
	done chan bool
}

var (
	impl Client
)

// Set defines the default progress client implementation
func Set(pi Client) {
	impl = pi
}

// MultiStep creates a new MultiStep implementation
func MultiStep(total int, format string, a ...interface{}) Progress {
	if impl == nil {
		panic("No progress implementation was configured. Use progress.Set() before using progress.")
	}

	desc := fmt.Sprintf(format, a...)
	prg := &BaseProgress{total: total}
	impl.Desc(desc)
	return prg
}

func runStepLoop(prg *Loop, dur time.Duration) {
	for {
		select {
		case <-prg.done:
			return
		default:
			impl.Step()
			time.Sleep(dur)
		}
	}
}

// NewLoop creates a new Loop based progress implementation
func NewLoop(format string, a ...interface{}) Progress {
	if impl == nil {
		panic("No progress implementation was configured. Use progress.Set() before using progress.")
	}

	desc := fmt.Sprintf(format, a...)
	prg := &Loop{}
	prg.done = make(chan bool)

	impl.Desc(desc)
	go runStepLoop(prg, impl.LoopWaitDuration())

	return prg
}

// Done notifies the actual implementation we have finished a task, this is the
// specific implementation for Loop based progress
func (prg *Loop) Done() {
	prg.done <- true
	impl.Done()
}

// Partial notifies the actual implementation we've moved one step on the
// set of steps for the MultiStep progress implementation
func (prg *BaseProgress) Partial(step int) {
	impl.Partial(prg.total, step)
}

// Done is the common BaseProgress implementation and simply notify the actual
// implementation we've finished a task
func (prg *BaseProgress) Done() {
	impl.Done()
}
