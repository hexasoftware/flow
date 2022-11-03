package flow

//
// Find a way to improve this mess, maybe it can be merged in one func
//
//

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"sync"
)

type executorFunc func(*Session, ...Data) (Data, error)

// Operation interface
type Operation interface { // Id perhaps?
	Process(params ...Data) (Data, error)
}

type operation struct {
	sync.Mutex
	flow     *Flow
	name     string
	kind     string
	inputs   []*operation // still figuring, might be Operation
	executor executorFunc // the executor?

	// Debug information for each operation
	file string
	line int
}

// NewOperation creates an operation
func (f *Flow) newOperation(kind string, inputs []*operation) *operation {
	_, file, line, _ := runtime.Caller(2) // outside of operation.go?
	return &operation{
		Mutex:  sync.Mutex{},
		flow:   f,
		kind:   kind,
		inputs: inputs,

		file: file,
		line: line,
		//name:   fmt.Sprintf("(var)<%s>", name),
	}

}
func (o *operation) String() string {
	_, file := path.Split(o.file)
	return fmt.Sprintf("[%s:%d]:{%s,%s}", file, o.line, o.kind, o.name)
}

// Process the operation with a new session
// ginputs are the global inputs
func (o *operation) Process(ginputs ...Data) (Data, error) {
	s := o.flow.NewSession()
	return s.run(o, ginputs...)
}

// Var create a operation
func (f *Flow) Var(name string, initial Data) Operation {
	inputs := f.makeInputs(initial)

	op := f.newOperation("var", inputs)
	op.executor = func(sess *Session, ginputs ...Data) (Data, error) {
		if name == "" {
			return nil, errors.New("Invalid name")
		}
		val, ok := f.Data.Load(name)
		if !ok {
			var initial Data
			res, err := sess.processInputs(op, ginputs...)
			if err != nil {
				return nil, err
			}

			if len(res) > 0 {
				initial = res[0]
			}

			val = initial
			f.Data.Store(name, val)
		}
		return val, nil
	}
	return op
}

// SetVar a variable from operation/constant
func (f *Flow) SetVar(name string, data Data) Operation {
	inputs := f.makeInputs(data)
	op := f.newOperation("setvar", inputs)
	op.executor = func(sess *Session, ginputs ...Data) (Data, error) {
		if name == "" {
			return nil, errors.New("Invalid name")
		}
		res, err := sess.processInputs(op, ginputs...)
		if err != nil {
			return nil, err
		}

		f.Data.Store(name, res[0])
		return res[0], nil
	}

	return op
}

// Op operation from registry
func (f *Flow) Op(name string, params ...interface{}) Operation {
	inputs := f.makeInputs(params...)

	// Grab executor here
	registryFn, err := f.registry.Get(name)
	if err != nil {
		return f.ErrOp(err)
	}
	op := f.newOperation("func", inputs)
	op.name = name
	// make executor from registry func
	op.executor = makeExecutor(op, registryFn)
	f.operations = append(f.operations, op)
	return op
}

// ErrOp define a nil operation that will return error
// Usefull for builders
func (f *Flow) ErrOp(err error) Operation {
	op := f.newOperation("error", nil)
	op.executor = func(*Session, ...Data) (Data, error) { return nil, err }
	return op
}

// Const define a const by defined ID
func (f *Flow) Const(value Data) Operation {
	// Optimize this definition
	constID := -1
	for k, v := range f.consts {
		if reflect.DeepEqual(v, value) {
			constID = k
			break
		}
	}
	if constID == -1 {
		constID = len(f.consts)
		f.consts = append(f.consts, value)
	}

	op := f.newOperation("const", nil)
	op.executor = func(*Session, ...Data) (Data, error) { return f.consts[constID], nil }
	return op
}

// In define input operation
func (f *Flow) In(paramID int) Operation {
	op := f.newOperation("in", nil)
	op.executor = func(sess *Session, ginputs ...Data) (Data, error) {
		if paramID < 0 || paramID >= len(ginputs) {
			return nil, ErrInput
		}
		return ginputs[paramID], nil
	}
	return op
}

func (f *Flow) makeInputs(params ...Data) []*operation {
	inputs := make([]*operation, len(params))
	for i, p := range params {
		switch v := p.(type) {
		case *operation:
			inputs[i] = v
		default:
			c := f.Const(v)
			inputs[i], _ = c.(*operation)
		}
	}
	return inputs
}

// make any go func as an executor
func makeExecutor(op *operation, fn interface{}) executorFunc {
	fnval := reflect.ValueOf(fn)
	callParam := make([]reflect.Value, fnval.Type().NumIn())

	// ExecutorFunc
	return func(sess *Session, ginputs ...Data) (Data, error) {
		// Change to wait to wait for the inputs
		inRes, err := sess.processInputs(op, ginputs...)
		if err != nil {
			return nil, err
		}
		// If the fn is a special we execute directly
		if gFn, ok := fn.(func(...Data) (Data, error)); ok {
			return gFn(inRes...)
		}

		for i, r := range inRes {
			if r == nil {
				callParam[i] = reflect.Zero(fnval.Type().In(i))
			} else {
				callParam[i] = reflect.ValueOf(r)
			}
		}

		// Start again and execute function
		fnret := fnval.Call(callParam)
		if len(fnret) == 0 {
			return nil, nil
		}
		// Output erroring
		if len(fnret) > 1 && (fnret[len(fnret)-1].Interface() != nil) {
			err, ok := fnret[len(fnret)-1].Interface().(error)
			if !ok {
				err = errors.New("unknown error")
			}
			return nil, err
		}

		// THE RESULT
		ret := fnret[0].Interface()
		return ret, nil
	}
}
