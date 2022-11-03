package flow

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

// Session operation session
type Session struct {
	*sync.Map
	flow    *Flow
	ginputs []Data
}

// NewSession creates a running context
func (f *Flow) NewSession() *Session {
	return &Session{
		Map:  &sync.Map{},
		flow: f,
	}
}

// Inputs sets the global graph inputs
func (s *Session) Inputs(ginputs ...Data) {
	s.ginputs = ginputs
}

// Run session run
func (s *Session) Run(ops ...Operation) ([]Data, error) {
	oplist := make([]*operation, len(ops))
	for i, op := range ops {
		oplist[i] = op.(*operation)
	}

	return s.goRunList(oplist, s.ginputs...)
}

// The main run function?
// run runs the operation
func (s *Session) run(op *operation, ginputs ...Data) (Data, error) {
	op.Lock()
	defer op.Unlock()
	// Load from cache if any
	if v, ok := s.Load(op); ok {
		return v, nil
	}

	res, err := s.triggerRun(op, ginputs...)
	if err != nil {
		return nil, err
	}

	s.Store(op, res)
	return res, nil

}

// RunList more than one operations in this session
func (s *Session) runList(ops []*operation, ginputs ...Data) ([]Data, error) {
	nOps := len(ops)
	// Total inputs
	callParam := make([]Data, nOps)
	// Parallel processing if inputs
	for i, op := range ops {
		res, err := s.run(op, ginputs...)
		if err != nil {
			return nil, err
		}
		callParam[i] = res
	}
	return callParam, nil
}

// Parallel
func (s *Session) goRunList(ops []*operation, ginputs ...Data) ([]Data, error) {
	nOps := len(ops)

	// Total inputs
	callParam := make([]Data, nOps)

	callErrors := ""
	paramMutex := sync.Mutex{}
	// Parallel processing if inputs
	wg := sync.WaitGroup{}
	wg.Add(nOps)
	for i, op := range ops {
		go func(i int, op *operation) {
			defer wg.Done()

			res, err := s.run(op, ginputs...)
			paramMutex.Lock()
			defer paramMutex.Unlock()
			if err != nil {
				callErrors += err.Error() + "\n"
				return
			}
			callParam[i] = res
		}(i, op)
	}
	wg.Wait()

	if callErrors != "" {
		return nil, errors.New(callErrors)
	}

	return callParam, nil
}

// make Executor for func
// safe run a func
//
func (s *Session) triggerRun(op *operation, ginputs ...Data) (Data, error) {
	s.flow.hooks.start(op)
	var err error
	var res Data

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "%v, %v\n", op, r)
				for i, in := range op.inputs {
					res, _ := s.run(in, ginputs...)
					fmt.Fprintf(os.Stderr, "Input: %d = %v\n", i, res)
				}
				debug.PrintStack()
				err = fmt.Errorf("%v %v", op, r)
			}
		}()
		res, err = op.executor(s, ginputs...)
	}()
	if err != nil {
		s.flow.hooks.error(op, err)
	} else {
		s.flow.hooks.finish(op, res)
	}
	return res, err

}
func (s *Session) processInputs(op *operation, ginputs ...Data) ([]Data, error) {
	s.flow.hooks.wait(op)
	res, err := s.goRunList(op.inputs, ginputs...)
	s.flow.hooks.start(op) // Back to start
	return res, err
}
